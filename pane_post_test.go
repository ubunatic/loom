package loom

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/term"
)

// postProbe keeps unguarded state: Post callbacks and Draw both touch it, so
// the race detector flags any callback that runs off the event-loop goroutine.
type postProbe struct {
	got   map[int][]int
	draws int
	// drawsAtFirstPost is the draw count seen by the first callback, -1 if none ran.
	drawsAtFirstPost int
	done             chan struct{}
}

func newPostProbe() *postProbe {
	return &postProbe{got: map[int][]int{}, drawsAtFirstPost: -1, done: make(chan struct{})}
}

func (p *postProbe) Draw(c *Canvas, _ Rect) {
	p.draws++
	c.Write(0, 0, "posts", Style{})
	for _, items := range p.got {
		_ = len(items)
	}
}
func (*postProbe) ConsumeMouse(MouseEvent) EventResult { return Ignored() }
func (p *postProbe) ConsumeKey(e KeyEvent) EventResult {
	if e.Key == "q" || e.Text == "q" {
		return QuitResult()
	}
	return Ignored()
}

func newPostTestPane(t *testing.T) (*Pane, func() error) {
	t.Helper()
	master, slave := openPTY(t)
	setPTYSize(t, master, 80, 24)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, wantRows: 1, cols: 80, startRow: 1, ResizeConfig: DefaultResizeConfig()}
	stop := make(chan struct{})
	t.Cleanup(func() { close(stop) })
	drainPTY(master, stop)
	quit := func() error {
		_, err := master.Write([]byte("q"))
		return err
	}
	return p, quit
}

func runPostPane(t *testing.T, p *Pane, root Widget) <-chan error {
	t.Helper()
	errch := make(chan error, 1)
	go func() { errch <- p.Run(root) }()
	return errch
}

func waitPostRun(t *testing.T, errch <-chan error) {
	t.Helper()
	select {
	case err := <-errch:
		if err != nil {
			t.Fatalf("Pane.Run: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("pane did not stop")
	}
}

func TestPanePostFromManyGoroutinesKeepsOrderOnLoop(t *testing.T) {
	const producers, perProducer = 20, 50 // 1000 posts
	p, quit := newPostTestPane(t)
	probe := newPostProbe()
	errch := runPostPane(t, p, probe)

	var wg sync.WaitGroup
	wg.Add(producers)
	for g := range producers {
		go func() {
			defer wg.Done()
			for i := range perProducer {
				p.Post(func() { probe.got[g] = append(probe.got[g], i) })
			}
		}()
	}
	wg.Wait()
	// FIFO: once this runs, every earlier post has run too.
	p.Post(func() { close(probe.done) })
	select {
	case <-probe.done:
	case <-time.After(3 * time.Second):
		t.Fatal("posted callbacks were not delivered")
	}
	if err := quit(); err != nil {
		t.Fatal(err)
	}
	waitPostRun(t, errch)

	total := 0
	for g := range producers {
		items := probe.got[g]
		total += len(items)
		for i, v := range items {
			if v != i {
				t.Fatalf("producer %d: item %d = %d, want FIFO order", g, i, v)
			}
		}
	}
	if total != producers*perProducer {
		t.Fatalf("delivered %d callbacks, want %d", total, producers*perProducer)
	}
}

func TestPanePostBeforeRunDeliveredOnFirstLoopTurn(t *testing.T) {
	p, quit := newPostTestPane(t)
	probe := newPostProbe()
	for i := range 3 {
		p.Post(func() {
			if i == 0 {
				probe.drawsAtFirstPost = probe.draws
			}
			probe.got[0] = append(probe.got[0], i)
		})
	}
	p.Post(func() { close(probe.done) })
	errch := runPostPane(t, p, probe)
	select {
	case <-probe.done:
	case <-time.After(3 * time.Second):
		t.Fatal("callbacks posted before Run were not delivered")
	}
	if err := quit(); err != nil {
		t.Fatal(err)
	}
	waitPostRun(t, errch)
	if probe.drawsAtFirstPost != 0 {
		t.Fatalf("first posted callback ran after %d draws, want before the first frame", probe.drawsAtFirstPost)
	}
	if got := probe.got[0]; len(got) != 3 || got[0] != 0 || got[2] != 2 {
		t.Fatalf("pre-run posts = %v, want [0 1 2]", got)
	}
}

func TestPanePostAfterCloseIsNoOp(t *testing.T) {
	_, slave := openPTY(t)
	p := &Pane{tty: slave, fd: int(slave.Fd()), rows: 1, cols: 40, startRow: 1}
	state, err := term.GetState(p.fd)
	if err != nil {
		t.Fatal(err)
	}
	p.oldState = state
	p.Post(func() {})
	p.Close()
	ran := false
	p.Post(func() { ran = true })
	if p.runPosted() || ran {
		t.Fatal("Post after Close queued or ran a callback")
	}
}
