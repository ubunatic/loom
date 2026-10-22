#!/usr/bin/env bash
set -euo pipefail

width=90
height=20
with_sgr=false
rows=
wait_text=
timeout_seconds=5
poll_seconds=0.1
output_dir=

usage() {
   printf 'Usage: %s [-x columns] [-y rows] [-e] [--rows first:last] --wait text [--timeout seconds] [--interval seconds] [--output-dir dir] -- command...\n' "$0" >&2
}

fail() {
   printf 'ERROR: %s\n' "$*" >&2
   exit 1
}

while test "$#" -gt 0
do
   case "$1" in
      -x)
         width="${2:?missing width}"
         shift 2
         ;;
      -y)
         height="${2:?missing height}"
         shift 2
         ;;
      -e)
         with_sgr=true
         shift
         ;;
      --rows)
         rows="${2:?missing row range}"
         shift 2
         ;;
      --wait)
         wait_text="${2:?missing wait text}"
         shift 2
         ;;
      --timeout)
         timeout_seconds="${2:?missing timeout}"
         shift 2
         ;;
      --interval)
         poll_seconds="${2:?missing interval}"
         shift 2
         ;;
      --output-dir)
         output_dir="${2:?missing output directory}"
         shift 2
         ;;
      --)
         shift
         break
         ;;
      -h|--help)
         usage
         exit 0
         ;;
      *)
         usage
         fail "unknown option: $1"
         ;;
   esac
done

if test "$#" -eq 0
then usage
     fail "command required after --"
fi
if test -z "$wait_text"
then fail "--wait is required"
fi
if ! printf '%s\n' "$width" | grep -Eq '^[0-9]+$' || ! printf '%s\n' "$height" | grep -Eq '^[0-9]+$' || test "$width" -lt 1 || test "$height" -lt 1
then fail "width and height must be positive integers"
fi
if ! printf '%s\n' "$timeout_seconds" | grep -Eq '^[0-9]+([.][0-9]+)?$' || ! printf '%s\n' "$poll_seconds" | grep -Eq '^[0-9]+([.][0-9]+)?$' || test "$(awk -v value="$timeout_seconds" 'BEGIN { print (value > 0) }')" != 1 || test "$(awk -v value="$poll_seconds" 'BEGIN { print (value > 0) }')" != 1
then fail "timeout and interval must be positive seconds"
fi
if test -n "$rows" && ! printf '%s\n' "$rows" | grep -Eq '^[1-9][0-9]*:[1-9][0-9]*$'
then fail "--rows must be first:last using 1-based rows"
fi
if ! command -v tmux >/dev/null 2>&1
then fail "tmux is required"
fi
if ! command -v loom >/dev/null 2>&1
then fail "installed loom command is required"
fi

socket="loom-live-$$"
session="live-$$"
temp_dir=$(mktemp -d)
cleanup() {
   tmux -f /dev/null -L "$socket" kill-server >/dev/null 2>&1 || true
   rm -rf "$temp_dir"
}
trap cleanup EXIT
if test -n "$output_dir"
then mkdir -p "$output_dir"
else output_dir=$(mktemp -d)
   printf 'snapshot directory: %s\n' "$output_dir"
fi

printf -v command_line '%q ' "$@"
tmux -f /dev/null -L "$socket" new-session -d -s "$session" -x "$width" -y "$height" "$command_line"
tmux -f /dev/null -L "$socket" set-option -s escape-time 0
printf 'tmux socket: %s\n' "$socket"

capture() {
   local mode="$1"
   if test "$mode" = sgr
   then tmux -f /dev/null -L "$socket" capture-pane -p -e -t "$session"
   else tmux -f /dev/null -L "$socket" capture-pane -p -t "$session"
   fi
}

contains_wait_text() {
   capture plain > "$temp_dir/wait.ansi"
   grep -Fq -- "$wait_text" "$temp_dir/wait.ansi"
}

wait_for_text() {
   local started now deadline timeout_ns
   started=$(date +%s%N)
   timeout_ns=$(awk -v seconds="$timeout_seconds" 'BEGIN { printf "%.0f", seconds * 1000000000 }')
   deadline=$((started + timeout_ns))
   while ! contains_wait_text
   do
      if ! tmux -f /dev/null -L "$socket" has-session -t "$session" 2>/dev/null
      then fail "command exited before rendering: $wait_text"
      fi
      now=$(date +%s%N)
      if test "$now" -ge "$deadline"
      then capture plain > "$temp_dir/timeout.ansi"
         printf 'ERROR: timed out waiting for pane text: %s; last pane follows:\n' "$wait_text" >&2
         cat "$temp_dir/timeout.ansi" >&2
         exit 1
      fi
      timeout "$poll_seconds" tail -f /dev/null || test "$?" -eq 124
   done
}

print_screen() {
   local mode="$1" text
   text=$(capture "$mode")
   if test -n "$rows"
   then local first="${rows%:*}" last="${rows#*:}"
      if test "$last" -lt "$first"
      then fail "--rows end must be >= start"
      fi
      text=$(printf '%s\n' "$text" | sed -n "${first},${last}p")
   fi
   printf '%s\n' "$text"
}

wait_for_text
step=0
while IFS= read -r line || test -n "$line"
do
   if test -z "$line"
   then continue
   fi
   read -r action rest <<< "$line"
   step=$((step + 1))
   case "$action" in
      keys)
         test -n "${rest:-}" || fail "step $step: keys requires tmux key names"
         read -r -a key_args <<< "$rest"
         tmux -f /dev/null -L "$socket" send-keys -t "$session" "${key_args[@]}"
         ;;
      hex)
         test -n "${rest:-}" || fail "step $step: hex requires byte pairs"
         read -r -a bytes <<< "$rest"
         for byte in "${bytes[@]}"
         do
            printf '%s\n' "$byte" | grep -Eq '^[[:xdigit:]]{2}$' || fail "step $step: invalid hex byte $byte"
         done
         tmux -f /dev/null -L "$socket" send-keys -H -t "$session" "${bytes[@]}"
         ;;
      wait)
         test -n "${rest:-}" || fail "step $step: wait requires text"
         wait_text="$rest"
         wait_for_text
         ;;
      pause)
         printf '%s\n' "${rest:-}" | grep -Eq '^[0-9]+([.][0-9]+)?$' || fail "step $step: pause requires seconds"
         timeout "$rest" tail -f /dev/null || test "$?" -eq 124
         ;;
      snap)
         label="${rest:-step-$step}"
         file="$output_dir/$(printf '%03d' "$step")-$label.ansi"
         capture sgr > "$file"
         printf 'snapshot: %s\n' "$file"
         ;;
      diff)
         test -n "${rest:-}" || fail "step $step: diff requires want.ansi"
         capture sgr > "$temp_dir/current.ansi"
         loom diff "$rest" "$temp_dir/current.ansi"
         ;;
      *)
         fail "step $step: unknown action $action"
         ;;
   esac
   if test "$action" != diff
   then if test "$with_sgr" = true
        then print_screen sgr
        else print_screen plain
      fi
   else
      print_screen plain
   fi
   printf 'step %d: %s\n' "$step" "$line"
done

