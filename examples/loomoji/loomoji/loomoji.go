// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package loomoji demonstrates a compact inline emoji and symbol picker.
package loomoji

import (
	"fmt"
	"os"
	"strings"

	"ubunatic.com/loom"
)

// ── Data model ────────────────────────────────────────────────────────────────

type entry struct {
	icon  string
	name  string
	group int
}

// category holds the display name and icon for a category tab.
type category struct {
	icon  string
	label string
}

// group indices — keep in sync with categories slice below.
const (
	grpFaces = iota
	grpHands
	grpAnimals
	grpFood
	grpSports
	grpTravel
	grpObjects
	grpHearts
	grpNature
	grpSymbols
	grpArrows
	grpBlocks
	grpLines
	grpBox
)

func categories() []category {
	return []category{
		{"😀", "Faces"},
		{"👋", "Hands"},
		{"🐾", "Animals"},
		{"🍔", "Food"},
		{"⚽", "Sports"},
		{"🚀", "Travel"},
		{"💡", "Objects"},
		{"❤️", "Hearts"},
		{"🌿", "Nature"},
		{"🔣", "Symbols"},
		{"➔", "Arrows"},
		{"█", "Blocks"},
		{"─", "Lines"},
		{"┼", "Box"},
	}
}

func entries() []entry {
	return entryList
}

var entryList = []entry{
	// ── Faces ─────────────────────────────────────────────────────────────────
	{"😀", "grinning face", grpFaces},
	{"😃", "grinning face big eyes", grpFaces},
	{"😄", "grinning face smiling eyes", grpFaces},
	{"😁", "beaming face", grpFaces},
	{"😆", "squinting face", grpFaces},
	{"😅", "grinning face sweat", grpFaces},
	{"😂", "face tears joy", grpFaces},
	{"🤣", "rolling floor laughing", grpFaces},
	{"😊", "smiling face smiling eyes", grpFaces},
	{"😇", "smiling face halo", grpFaces},
	{"🙂", "slightly smiling face", grpFaces},
	{"🙃", "upside down face", grpFaces},
	{"😉", "winking face", grpFaces},
	{"😍", "smiling face heart eyes", grpFaces},
	{"🥰", "smiling face hearts", grpFaces},
	{"😘", "face blowing kiss", grpFaces},
	{"😋", "face savoring food", grpFaces},
	{"😛", "face tongue", grpFaces},
	{"😜", "winking face tongue", grpFaces},
	{"🤪", "zany face", grpFaces},
	{"😝", "squinting face tongue", grpFaces},
	{"🤗", "hugging face", grpFaces},
	{"🤔", "thinking face", grpFaces},
	{"🫡", "saluting face", grpFaces},
	{"🤨", "face raised eyebrow", grpFaces},
	{"😐", "neutral face", grpFaces},
	{"😑", "expressionless face", grpFaces},
	{"😶", "face without mouth", grpFaces},
	{"🙄", "face rolling eyes", grpFaces},
	{"😏", "smirking face", grpFaces},
	{"😎", "smiling face sunglasses", grpFaces},
	{"🥳", "partying face", grpFaces},
	{"😭", "loudly crying face", grpFaces},
	{"😢", "crying face", grpFaces},
	{"😴", "sleeping face", grpFaces},
	{"🤤", "drooling face", grpFaces},
	{"😪", "sleepy face", grpFaces},
	{"😮‍💨", "face exhaling", grpFaces},
	{"😵", "knocked out face", grpFaces},
	{"😵‍💫", "face spiral eyes", grpFaces},
	{"🤯", "exploding head", grpFaces},
	{"🥱", "yawning face", grpFaces},
	{"😤", "face steam nose", grpFaces},
	{"😡", "pouting face", grpFaces},
	{"😠", "angry face", grpFaces},
	{"🤬", "face symbols mouth", grpFaces},
	{"😈", "smiling face horns", grpFaces},
	{"👿", "angry face horns", grpFaces},
	{"💀", "skull", grpFaces},
	{"☠️", "skull crossbones", grpFaces},
	{"💩", "pile poo", grpFaces},
	{"🤡", "clown face", grpFaces},
	{"👹", "ogre", grpFaces},
	{"👺", "goblin", grpFaces},
	{"👻", "ghost", grpFaces},
	{"👽", "alien", grpFaces},
	{"👾", "alien monster", grpFaces},
	{"🤖", "robot", grpFaces},
	{"😺", "grinning cat", grpFaces},
	{"😸", "grinning cat smiling eyes", grpFaces},
	{"😹", "cat tears joy", grpFaces},
	{"😻", "smiling cat heart eyes", grpFaces},
	{"😼", "cat wry smile", grpFaces},
	{"😽", "kissing cat", grpFaces},
	{"🙀", "weary cat", grpFaces},
	{"😿", "crying cat", grpFaces},
	{"😾", "pouting cat", grpFaces},
	{"🙈", "see no evil monkey", grpFaces},
	{"🙉", "hear no evil monkey", grpFaces},
	{"🙊", "speak no evil monkey", grpFaces},
	{"🤩", "star struck", grpFaces},
	{"🥺", "pleading face", grpFaces},
	{"😬", "grimacing face", grpFaces},
	{"🫠", "melting face", grpFaces},
	{"🥸", "disguised face", grpFaces},
	{"🥶", "cold face", grpFaces},
	{"🥵", "hot face", grpFaces},

	// ── Hands ─────────────────────────────────────────────────────────────────
	{"👋", "waving hand", grpHands},
	{"🤚", "raised back hand", grpHands},
	{"🖐️", "hand fingers splayed", grpHands},
	{"✋", "raised hand", grpHands},
	{"🖖", "vulcan salute", grpHands},
	{"👌", "ok hand", grpHands},
	{"🤌", "pinched fingers", grpHands},
	{"🤏", "pinching hand", grpHands},
	{"✌️", "victory hand", grpHands},
	{"🤞", "crossed fingers", grpHands},
	{"🫰", "hand index thumb crossed", grpHands},
	{"🤟", "love you gesture", grpHands},
	{"🤘", "sign of horns", grpHands},
	{"🤙", "call me hand", grpHands},
	{"👈", "pointing left", grpHands},
	{"👉", "pointing right", grpHands},
	{"👆", "pointing up", grpHands},
	{"🖕", "middle finger", grpHands},
	{"👇", "pointing down", grpHands},
	{"☝️", "index pointing up", grpHands},
	{"👍", "thumbs up", grpHands},
	{"👎", "thumbs down", grpHands},
	{"✊", "raised fist", grpHands},
	{"👊", "oncoming fist", grpHands},
	{"🤛", "left facing fist", grpHands},
	{"🤜", "right facing fist", grpHands},
	{"👏", "clapping hands", grpHands},
	{"🙌", "raising hands", grpHands},
	{"👐", "open hands", grpHands},
	{"🤲", "palms up together", grpHands},
	{"🤝", "handshake", grpHands},
	{"🙏", "folded hands", grpHands},
	{"✍️", "writing hand", grpHands},
	{"💅", "nail polish", grpHands},
	{"🤳", "selfie", grpHands},
	{"💪", "flexed biceps", grpHands},
	{"🦾", "mechanical arm", grpHands},
	{"🦿", "mechanical leg", grpHands},
	{"🦵", "leg", grpHands},
	{"🦶", "foot", grpHands},
	{"👂", "ear", grpHands},
	{"🦻", "ear hearing aid", grpHands},
	{"👃", "nose", grpHands},
	{"🧠", "brain", grpHands},
	{"🫀", "anatomical heart", grpHands},
	{"🫁", "lungs", grpHands},
	{"🦷", "tooth", grpHands},
	{"🦴", "bone", grpHands},
	{"👀", "eyes", grpHands},
	{"👁️", "eye", grpHands},
	{"👅", "tongue", grpHands},
	{"👄", "mouth", grpHands},

	// ── Animals ───────────────────────────────────────────────────────────────
	{"🐶", "dog face", grpAnimals},
	{"🐱", "cat face", grpAnimals},
	{"🐭", "mouse face", grpAnimals},
	{"🐹", "hamster face", grpAnimals},
	{"🐰", "rabbit face", grpAnimals},
	{"🦊", "fox", grpAnimals},
	{"🐻", "bear", grpAnimals},
	{"🐼", "panda", grpAnimals},
	{"🐻‍❄️", "polar bear", grpAnimals},
	{"🐨", "koala", grpAnimals},
	{"🐯", "tiger face", grpAnimals},
	{"🦁", "lion", grpAnimals},
	{"🐮", "cow face", grpAnimals},
	{"🐷", "pig face", grpAnimals},
	{"🐽", "pig nose", grpAnimals},
	{"🐸", "frog", grpAnimals},
	{"🐵", "monkey face", grpAnimals},
	{"🐒", "monkey", grpAnimals},
	{"🐔", "chicken", grpAnimals},
	{"🐧", "penguin", grpAnimals},
	{"🐦", "bird", grpAnimals},
	{"🐤", "baby chick", grpAnimals},
	{"🐣", "hatching chick", grpAnimals},
	{"🐥", "front chick", grpAnimals},
	{"🦆", "duck", grpAnimals},
	{"🦅", "eagle", grpAnimals},
	{"🦉", "owl", grpAnimals},
	{"🦇", "bat", grpAnimals},
	{"🐺", "wolf", grpAnimals},
	{"🐗", "boar", grpAnimals},
	{"🐴", "horse face", grpAnimals},
	{"🦄", "unicorn", grpAnimals},
	{"🐝", "honeybee", grpAnimals},
	{"🪱", "worm", grpAnimals},
	{"🐛", "bug caterpillar", grpAnimals},
	{"🦋", "butterfly", grpAnimals},
	{"🐌", "snail", grpAnimals},
	{"🐞", "lady beetle", grpAnimals},
	{"🐜", "ant", grpAnimals},
	{"🪰", "fly", grpAnimals},
	{"🪲", "beetle", grpAnimals},
	{"🪳", "cockroach", grpAnimals},
	{"🦟", "mosquito", grpAnimals},
	{"🦗", "cricket", grpAnimals},
	{"🕷️", "spider", grpAnimals},
	{"🕸️", "spider web", grpAnimals},
	{"🦂", "scorpion", grpAnimals},
	{"🐢", "turtle", grpAnimals},
	{"🐍", "snake", grpAnimals},
	{"🦎", "lizard", grpAnimals},
	{"🦖", "t-rex dinosaur", grpAnimals},
	{"🦕", "sauropod dinosaur", grpAnimals},
	{"🐙", "octopus", grpAnimals},
	{"🦑", "squid", grpAnimals},
	{"🦐", "shrimp", grpAnimals},
	{"🦞", "lobster", grpAnimals},
	{"🦀", "crab", grpAnimals},
	{"🐡", "blowfish", grpAnimals},
	{"🐠", "tropical fish", grpAnimals},
	{"🐟", "fish", grpAnimals},
	{"🐬", "dolphin", grpAnimals},
	{"🐳", "spouting whale", grpAnimals},
	{"🐋", "whale", grpAnimals},
	{"🦈", "shark", grpAnimals},
	{"🐊", "crocodile", grpAnimals},
	{"🐅", "tiger", grpAnimals},
	{"🐆", "leopard", grpAnimals},
	{"🦓", "zebra", grpAnimals},
	{"🦍", "gorilla", grpAnimals},
	{"🦧", "orangutan", grpAnimals},
	{"🦣", "mammoth", grpAnimals},
	{"🐘", "elephant", grpAnimals},
	{"🦛", "hippopotamus", grpAnimals},
	{"🦏", "rhinoceros", grpAnimals},
	{"🐪", "camel", grpAnimals},
	{"🐫", "two-hump camel", grpAnimals},
	{"🦒", "giraffe", grpAnimals},
	{"🦘", "kangaroo", grpAnimals},
	{"🦬", "bison", grpAnimals},
	{"🐃", "water buffalo", grpAnimals},
	{"🐂", "ox", grpAnimals},
	{"🐄", "cow", grpAnimals},
	{"🐎", "horse", grpAnimals},
	{"🐖", "pig", grpAnimals},
	{"🐏", "ram", grpAnimals},
	{"🐑", "sheep", grpAnimals},
	{"🦙", "llama", grpAnimals},
	{"🐐", "goat", grpAnimals},
	{"🦌", "deer", grpAnimals},
	{"🐕", "dog", grpAnimals},
	{"🐩", "poodle", grpAnimals},
	{"🦮", "guide dog", grpAnimals},
	{"🐈", "cat", grpAnimals},
	{"🐈‍⬛", "black cat", grpAnimals},
	{"🪶", "feather", grpAnimals},
	{"🐓", "rooster", grpAnimals},
	{"🦃", "turkey", grpAnimals},
	{"🦤", "dodo", grpAnimals},
	{"🦚", "peacock", grpAnimals},
	{"🦜", "parrot", grpAnimals},
	{"🦢", "swan", grpAnimals},
	{"🦩", "flamingo", grpAnimals},
	{"🕊️", "dove", grpAnimals},
	{"🐇", "rabbit", grpAnimals},
	{"🦝", "raccoon", grpAnimals},
	{"🦨", "skunk", grpAnimals},
	{"🦡", "badger", grpAnimals},
	{"🦫", "beaver", grpAnimals},
	{"🦦", "otter", grpAnimals},
	{"🦥", "sloth", grpAnimals},
	{"🐁", "mouse", grpAnimals},
	{"🐀", "rat", grpAnimals},
	{"🐿️", "chipmunk", grpAnimals},
	{"🦔", "hedgehog", grpAnimals},
	{"🐉", "dragon", grpAnimals},
	{"🐲", "dragon face", grpAnimals},

	// ── Food ──────────────────────────────────────────────────────────────────
	{"🍏", "green apple", grpFood},
	{"🍎", "red apple", grpFood},
	{"🍐", "pear", grpFood},
	{"🍊", "tangerine orange", grpFood},
	{"🍋", "lemon", grpFood},
	{"🍌", "banana", grpFood},
	{"🍉", "watermelon", grpFood},
	{"🍇", "grapes", grpFood},
	{"🍓", "strawberry", grpFood},
	{"🫐", "blueberries", grpFood},
	{"🍈", "melon", grpFood},
	{"🍒", "cherries", grpFood},
	{"🍑", "peach", grpFood},
	{"🥭", "mango", grpFood},
	{"🍍", "pineapple", grpFood},
	{"🥥", "coconut", grpFood},
	{"🥝", "kiwi fruit", grpFood},
	{"🍅", "tomato", grpFood},
	{"🍆", "eggplant aubergine", grpFood},
	{"🥑", "avocado", grpFood},
	{"🥦", "broccoli", grpFood},
	{"🥬", "leafy green", grpFood},
	{"🥒", "cucumber", grpFood},
	{"🌶️", "hot pepper chili", grpFood},
	{"🫑", "bell pepper", grpFood},
	{"🌽", "corn", grpFood},
	{"🥕", "carrot", grpFood},
	{"🫒", "olive", grpFood},
	{"🧄", "garlic", grpFood},
	{"🧅", "onion", grpFood},
	{"🥔", "potato", grpFood},
	{"🍠", "roasted sweet potato", grpFood},
	{"🥐", "croissant", grpFood},
	{"🥯", "bagel", grpFood},
	{"🍞", "bread", grpFood},
	{"🥖", "baguette bread", grpFood},
	{"🥨", "pretzel", grpFood},
	{"🧀", "cheese wedge", grpFood},
	{"🥚", "egg", grpFood},
	{"🍳", "cooking fried egg", grpFood},
	{"🧈", "butter", grpFood},
	{"🥞", "pancakes", grpFood},
	{"🧇", "waffle", grpFood},
	{"🥓", "bacon", grpFood},
	{"🥩", "cut of meat steak", grpFood},
	{"🍗", "poultry leg chicken", grpFood},
	{"🍖", "meat on bone", grpFood},
	{"🌭", "hot dog", grpFood},
	{"🍔", "hamburger burger", grpFood},
	{"🍟", "french fries", grpFood},
	{"🍕", "pizza", grpFood},
	{"🫓", "flatbread", grpFood},
	{"🥪", "sandwich", grpFood},
	{"🥙", "stuffed flatbread pita", grpFood},
	{"🧆", "falafel", grpFood},
	{"🌮", "taco", grpFood},
	{"🌯", "burrito", grpFood},
	{"🫔", "tamale", grpFood},
	{"🥗", "green salad", grpFood},
	{"🥘", "shallow pan food paella", grpFood},
	{"🫕", "fondue", grpFood},
	{"🥫", "canned food", grpFood},
	{"🍝", "spaghetti pasta", grpFood},
	{"🍜", "steaming bowl ramen noodles", grpFood},
	{"🍲", "pot of food stew", grpFood},
	{"🍛", "curry rice", grpFood},
	{"🍣", "sushi", grpFood},
	{"🍱", "bento box", grpFood},
	{"🥟", "dumpling", grpFood},
	{"🦪", "oyster", grpFood},
	{"🍤", "fried shrimp tempura", grpFood},
	{"🍙", "rice ball onigiri", grpFood},
	{"🍚", "cooked rice", grpFood},
	{"🍘", "rice cracker senbei", grpFood},
	{"🍥", "fish cake swirl narutomaki", grpFood},
	{"🥠", "fortune cookie", grpFood},
	{"🥮", "moon cake", grpFood},
	{"🍢", "oden", grpFood},
	{"🍡", "dango", grpFood},
	{"🍧", "shaved ice", grpFood},
	{"🍨", "ice cream", grpFood},
	{"🍦", "soft ice cream cone", grpFood},
	{"🥧", "pie", grpFood},
	{"🧁", "cupcake", grpFood},
	{"🍰", "shortcake strawberry cake", grpFood},
	{"🎂", "birthday cake", grpFood},
	{"🍮", "custard flan pudding", grpFood},
	{"🍭", "lollipop", grpFood},
	{"🍬", "candy", grpFood},
	{"🍫", "chocolate bar", grpFood},
	{"🍿", "popcorn", grpFood},
	{"🍩", "doughnut donut", grpFood},
	{"🍪", "cookie biscuit", grpFood},
	{"🌰", "chestnut", grpFood},
	{"🥜", "peanuts", grpFood},
	{"🍯", "honey pot", grpFood},
	{"🥛", "glass of milk", grpFood},
	{"🍼", "baby bottle", grpFood},
	{"☕", "hot beverage coffee tea", grpFood},
	{"🫖", "teapot", grpFood},
	{"🍵", "teacup without handle matcha", grpFood},
	{"🧃", "beverage box juice", grpFood},
	{"🥤", "cup with straw soda", grpFood},
	{"🧋", "bubble tea boba", grpFood},
	{"🍶", "sake", grpFood},
	{"🍺", "beer mug", grpFood},
	{"🍻", "clinking beer mugs", grpFood},
	{"🥂", "clinking glasses champagne", grpFood},
	{"🍷", "wine glass", grpFood},
	{"🥃", "tumbler glass whiskey", grpFood},
	{"🍸", "cocktail glass martini", grpFood},
	{"🍹", "tropical drink", grpFood},
	{"🧉", "mate", grpFood},
	{"🍾", "bottle with popping cork champagne", grpFood},
	{"🧊", "ice cube", grpFood},

	// ── Sports & Activities ───────────────────────────────────────────────────
	{"⚽", "soccer ball football", grpSports},
	{"🏀", "basketball", grpSports},
	{"🏈", "american football", grpSports},
	{"⚾", "baseball", grpSports},
	{"🥎", "softball", grpSports},
	{"🎾", "tennis", grpSports},
	{"🏐", "volleyball", grpSports},
	{"🏉", "rugby football", grpSports},
	{"🥏", "flying disc frisbee", grpSports},
	{"🎱", "pool 8 ball billiard", grpSports},
	{"🪀", "yo-yo", grpSports},
	{"🏓", "ping pong table tennis", grpSports},
	{"🏸", "badminton", grpSports},
	{"🏒", "ice hockey", grpSports},
	{"🏑", "field hockey", grpSports},
	{"🥍", "lacrosse", grpSports},
	{"🏏", "cricket game", grpSports},
	{"🪃", "boomerang", grpSports},
	{"🥅", "goal net", grpSports},
	{"⛳", "flag in hole golf", grpSports},
	{"🪁", "kite", grpSports},
	{"🏹", "bow and arrow archery", grpSports},
	{"🎣", "fishing pole", grpSports},
	{"🤿", "diving mask snorkel", grpSports},
	{"🥊", "boxing glove", grpSports},
	{"🥋", "martial arts uniform judo karate", grpSports},
	{"🎽", "running shirt sash", grpSports},
	{"🛹", "skateboard", grpSports},
	{"🛼", "roller skate", grpSports},
	{"🛷", "sled", grpSports},
	{"⛸️", "ice skate", grpSports},
	{"🥌", "curling stone", grpSports},
	{"🎿", "skis skiing", grpSports},
	{"⛷️", "skier", grpSports},
	{"🏂", "snowboarder", grpSports},
	{"🪂", "parachute", grpSports},
	{"🏋️", "person lifting weights weightlifter", grpSports},
	{"🤼", "people wrestling wrestlers", grpSports},
	{"🤸", "person cartwheeling gymnastics", grpSports},
	{"⛹️", "person bouncing ball basketball", grpSports},
	{"🤺", "person fencing fencer", grpSports},
	{"🤾", "person playing handball", grpSports},
	{"🏌️", "person golfing golfer", grpSports},
	{"🏇", "horse racing jockey", grpSports},
	{"🧘", "person in lotus position yoga meditation", grpSports},
	{"🏄", "person surfing surfer", grpSports},
	{"🏊", "person swimming swimmer", grpSports},
	{"🤽", "person playing water polo", grpSports},
	{"🚣", "person rowing boat", grpSports},
	{"🧗", "person climbing rock climber", grpSports},
	{"🚵", "person mountain biking", grpSports},
	{"🚴", "person biking cyclist", grpSports},
	{"🏆", "trophy champion winner", grpSports},
	{"🥇", "1st place medal gold", grpSports},
	{"🥈", "2nd place medal silver", grpSports},
	{"🥉", "3rd place medal bronze", grpSports},
	{"🏅", "sports medal", grpSports},
	{"🎖️", "military medal", grpSports},
	{"🏵️", "rosette", grpSports},
	{"🎗️", "reminder ribbon", grpSports},
	{"🎫", "ticket", grpSports},
	{"🎟️", "admission tickets", grpSports},
	{"🎪", "circus tent", grpSports},
	{"🤹", "person juggling juggler", grpSports},
	{"🎭", "performing arts theater masks", grpSports},
	{"🩰", "ballet shoes dancer", grpSports},
	{"🎨", "artist palette paint", grpSports},
	{"🎬", "clapper board cinema movie film", grpSports},
	{"🎤", "microphone singing karaoke", grpSports},
	{"🎧", "headphone audio music", grpSports},
	{"🎼", "musical score music notes", grpSports},
	{"🎹", "musical keyboard piano", grpSports},
	{"🥁", "drum", grpSports},
	{"🪘", "long drum djembe", grpSports},
	{"🎷", "saxophone jazz", grpSports},
	{"🎺", "trumpet brass", grpSports},
	{"🪗", "accordion", grpSports},
	{"🎸", "guitar rock music", grpSports},
	{"🪕", "banjo", grpSports},
	{"🎻", "violin classical music", grpSports},
	{"🎲", "game die dice board game", grpSports},
	{"♟️", "chess pawn chess board", grpSports},
	{"🎯", "bullseye target dart", grpSports},
	{"🎳", "bowling pins ball", grpSports},
	{"🎮", "video game controller joystick", grpSports},
	{"🎰", "slot machine casino gambling", grpSports},
	{"🧩", "puzzle piece jigsaw", grpSports},

	// ── Travel & Places ───────────────────────────────────────────────────────
	{"🚗", "automobile car", grpTravel},
	{"🚕", "taxi cab", grpTravel},
	{"🚙", "sport utility vehicle suv", grpTravel},
	{"🚌", "bus transit", grpTravel},
	{"🚎", "trolleybus", grpTravel},
	{"🏎️", "racing car formula 1", grpTravel},
	{"🚓", "police car", grpTravel},
	{"🚑", "ambulance emergency", grpTravel},
	{"🚒", "fire engine firetruck", grpTravel},
	{"🚐", "minibus van", grpTravel},
	{"🛻", "pickup truck", grpTravel},
	{"🚚", "delivery truck lorry", grpTravel},
	{"🚛", "articulated lorry semi truck", grpTravel},
	{"🚜", "tractor farm", grpTravel},
	{"🦯", "white cane blind", grpTravel},
	{"🦽", "manual wheelchair", grpTravel},
	{"🦼", "motorized wheelchair", grpTravel},
	{"🛴", "kick scooter", grpTravel},
	{"🚲", "bicycle bike", grpTravel},
	{"🛵", "motor scooter moped", grpTravel},
	{"🏍️", "motorcycle motorbike", grpTravel},
	{"🛺", "auto rickshaw tuk tuk", grpTravel},
	{"🚨", "police car light siren beacon", grpTravel},
	{"🚔", "oncoming police car", grpTravel},
	{"🚍", "oncoming bus", grpTravel},
	{"🚘", "oncoming automobile car", grpTravel},
	{"🚖", "oncoming taxi", grpTravel},
	{"🚡", "aerial tramway cable car", grpTravel},
	{"🚠", "mountain cableway", grpTravel},
	{"🚟", "suspension railway", grpTravel},
	{"🚃", "railway car train", grpTravel},
	{"🚋", "tram car streetcar", grpTravel},
	{"🚞", "mountain railway", grpTravel},
	{"🚝", "monorail", grpTravel},
	{"🚄", "high-speed train bullet train", grpTravel},
	{"🚅", "bullet train shinkansen", grpTravel},
	{"🚈", "light rail", grpTravel},
	{"🚂", "locomotive steam train", grpTravel},
	{"🚆", "train", grpTravel},
	{"🚇", "metro subway underground", grpTravel},
	{"🚊", "tram streetcar", grpTravel},
	{"🚉", "station train station", grpTravel},
	{"✈️", "airplane flight plane travel", grpTravel},
	{"🛫", "airplane departure takeoff", grpTravel},
	{"🛬", "airplane arrival landing", grpTravel},
	{"🛩️", "small airplane private jet", grpTravel},
	{"💺", "seat airline chair", grpTravel},
	{"🛰️", "satellite orbit space", grpTravel},
	{"🚀", "rocket spaceship launch", grpTravel},
	{"🛸", "flying saucer ufo alien", grpTravel},
	{"🚁", "helicopter chopper", grpTravel},
	{"🛶", "canoe kayak boat", grpTravel},
	{"⛵", "sailboat yacht", grpTravel},
	{"🚤", "speedboat motorboat", grpTravel},
	{"🛥️", "motor boat", grpTravel},
	{"🛳️", "passenger ship cruise liner", grpTravel},
	{"⛴️", "ferry boat", grpTravel},
	{"🚢", "ship boat cargo", grpTravel},
	{"⚓", "anchor port harbor", grpTravel},
	{"⛽", "fuel pump gas station petrol", grpTravel},
	{"🚧", "construction barrier sign", grpTravel},
	{"🚦", "vertical traffic light signal", grpTravel},
	{"🚥", "horizontal traffic light", grpTravel},
	{"🚏", "bus stop sign", grpTravel},
	{"🗺️", "world map atlas navigation", grpTravel},
	{"🗿", "moai easter island statue", grpTravel},
	{"🗽", "statue of liberty new york", grpTravel},
	{"🗼", "tokyo tower japan", grpTravel},
	{"🏰", "castle fairytale", grpTravel},
	{"🏯", "japanese castle", grpTravel},
	{"🏟️", "stadium arena", grpTravel},
	{"🎡", "ferris wheel amusement park", grpTravel},
	{"🎢", "roller coaster carnival", grpTravel},
	{"🎠", "carousel horse merry go round", grpTravel},
	{"⛲", "fountain water", grpTravel},
	{"🏖️", "beach with umbrella vacation", grpTravel},
	{"🏝️", "desert island tropical", grpTravel},
	{"🏜️", "desert dunes sahara", grpTravel},
	{"🌋", "volcano eruption lava", grpTravel},
	{"⛰️", "mountain peak nature", grpTravel},
	{"🏔️", "snow capped mountain alps", grpTravel},
	{"🗻", "mount fuji japan", grpTravel},
	{"🏕️", "camping campsite tent", grpTravel},
	{"⛺", "tent camp shelter", grpTravel},
	{"🏠", "house home residence", grpTravel},
	{"🏡", "house with garden suburban", grpTravel},
	{"🏘️", "houses neighborhood village", grpTravel},
	{"🏚️", "derelict house abandoned", grpTravel},
	{"🏗️", "building construction crane", grpTravel},
	{"🏭", "factory industry plant", grpTravel},
	{"🏢", "office building skyscraper", grpTravel},
	{"🏬", "department store mall shop", grpTravel},
	{"🏣", "japanese post office", grpTravel},
	{"🏤", "post office mail", grpTravel},
	{"🏥", "hospital medical clinic", grpTravel},
	{"🏦", "bank finance money", grpTravel},
	{"🏨", "hotel lodging resort", grpTravel},
	{"🏪", "convenience store 24 7 shop", grpTravel},
	{"🏫", "school university education", grpTravel},
	{"🏩", "love hotel", grpTravel},
	{"💒", "wedding chapel church", grpTravel},
	{"🏛️", "classical building museum government", grpTravel},
	{"⛪", "church christian cathedral", grpTravel},
	{"🕌", "mosque islamic muslim", grpTravel},
	{"🛕", "hindu temple", grpTravel},
	{"🕍", "synagogue jewish temple", grpTravel},
	{"⛩️", "shinto shrine torii gate", grpTravel},
	{"🕋", "kaaba mecca islam", grpTravel},

	// ── Objects ───────────────────────────────────────────────────────────────
	{"💡", "light bulb idea lamp", grpObjects},
	{"🔦", "flashlight torch light", grpObjects},
	{"🏮", "red paper lantern izakaya", grpObjects},
	{"🪔", "diya lamp oil", grpObjects},
	{"🧱", "brick masonry wall", grpObjects},
	{"🪵", "wood log timber lumber", grpObjects},
	{"🪙", "coin currency money", grpObjects},
	{"💰", "money bag cash wealth", grpObjects},
	{"💵", "dollar banknote cash money", grpObjects},
	{"💳", "credit card payment visa", grpObjects},
	{"💎", "gem stone diamond jewel precious", grpObjects},
	{"⚖️", "balance scale justice law legal", grpObjects},
	{"🪜", "ladder climb step", grpObjects},
	{"🧰", "toolbox tools kit", grpObjects},
	{"🪛", "screwdriver tool repair", grpObjects},
	{"🔧", "wrench spanner tool fix", grpObjects},
	{"🔨", "hammer tool build construct", grpObjects},
	{"⚒️", "hammer and pick mine mining", grpObjects},
	{"🛠️", "hammer and wrench tools fix repair", grpObjects},
	{"⛏️", "pickaxe mining gold minecraft", grpObjects},
	{"🪚", "carpentry saw woodwork hand tool", grpObjects},
	{"🔩", "nut and bolt screw hardware", grpObjects},
	{"⚙️", "gear cog settings config mechanics", grpObjects},
	{"🧲", "magnet attraction magnetic", grpObjects},
	{"🔫", "water pistol squirt gun weapon", grpObjects},
	{"💣", "bomb explosive dynamite tnt", grpObjects},
	{"🧨", "firecracker fireworks explosive", grpObjects},
	{"🪓", "axe woodchopper chop weapon", grpObjects},
	{"🔪", "kitchen knife chef blade cook", grpObjects},
	{"🗡️", "dagger blade knife sword weapon", grpObjects},
	{"⚔️", "crossed swords battle duel combat", grpObjects},
	{"🛡️", "shield defense protect security", grpObjects},
	{"🚬", "cigarette smoke smoking tobacco", grpObjects},
	{"⚰️", "coffin funeral death casket", grpObjects},
	{"🪦", "headstone tombstone grave cemetery", grpObjects},
	{"⚱️", "funeral urn ashes memorial", grpObjects},
	{"🏺", "amphora vase jar ancient pottery", grpObjects},
	{"🔮", "crystal ball fortune teller magic future", grpObjects},
	{"🧿", "nazar amulet evil eye turkish", grpObjects},
	{"💈", "barber pole haircut styling", grpObjects},
	{"⚗️", "alembic chemistry laboratory distillation", grpObjects},
	{"🧪", "test tube chemistry science experiment", grpObjects},
	{"🧫", "petri dish biology culture bacteria", grpObjects},
	{"🧬", "dna genetics biology molecule", grpObjects},
	{"🔬", "microscope science research biology", grpObjects},
	{"🔭", "telescope astronomy stars space", grpObjects},
	{"📡", "satellite antenna dish communication radar", grpObjects},
	{"💉", "syringe vaccine injection medical", grpObjects},
	{"🩹", "adhesive bandage bandaid injury first aid", grpObjects},
	{"🩺", "stethoscope doctor medical exam", grpObjects},
	{"🚪", "door entryway exit open", grpObjects},
	{"🪞", "mirror reflection glass", grpObjects},
	{"🪟", "window view glass architecture", grpObjects},
	{"🛏️", "bed sleep hotel furniture bedroom", grpObjects},
	{"🛋️", "couch and lamp sofa living room furniture", grpObjects},
	{"🪑", "chair seat furniture sitting", grpObjects},
	{"🚽", "toilet restroom bathroom wc", grpObjects},
	{"🪠", "plunger clog toilet plumbing", grpObjects},
	{"🚿", "shower bathroom water wash", grpObjects},
	{"🛁", "bathtub bathing hygiene soak", grpObjects},
	{"🧹", "broom sweep clean tidy witch", grpObjects},
	{"🧺", "basket laundry storage picnic", grpObjects},
	{"🧻", "roll of paper toilet paper tissue", grpObjects},
	{"🪣", "bucket pail water cleaning", grpObjects},
	{"🧼", "soap bar lather wash bubbles hygiene", grpObjects},
	{"🫧", "bubbles soap fizzy carbonated", grpObjects},
	{"🪥", "toothbrush dental hygiene teeth clean", grpObjects},
	{"🧽", "sponge cleaning wash kitchen", grpObjects},
	{"🧯", "fire extinguisher safety emergency rescue", grpObjects},
	{"🛒", "shopping cart trolley market store checkout", grpObjects},
	{"📦", "package box cardboard parcel delivery shipping", grpObjects},
	{"📫", "closed mailbox with raised flag mail letter", grpObjects},
	{"📪", "closed mailbox with lowered flag post", grpObjects},
	{"📬", "open mailbox with raised flag incoming", grpObjects},
	{"📭", "open mailbox with lowered flag empty", grpObjects},
	{"📮", "postbox mailbox royal mail letter", grpObjects},
	{"✉️", "envelope email letter mail message", grpObjects},
	{"📧", "e-mail electronic mail message inbox", grpObjects},
	{"📨", "incoming envelope mail message received", grpObjects},
	{"📩", "envelope with arrow send outgoing", grpObjects},
	{"📤", "outbox tray outgoing sent upload", grpObjects},
	{"📥", "inbox tray incoming received download", grpObjects},
	{"🏷️", "label tag price sale marker metadata", grpObjects},
	{"📁", "file folder directory organizing files", grpObjects},
	{"📂", "open file folder directory documents", grpObjects},
	{"🗂️", "card index dividers organizing tabs cards", grpObjects},
	{"📅", "calendar date schedule day year", grpObjects},
	{"📆", "tear-off calendar date appointment planner", grpObjects},
	{"🗒️", "spiral notepad memo notes pad", grpObjects},
	{"🗓️", "spiral calendar schedule planner agenda", grpObjects},
	{"📇", "card index rolodex contacts address book", grpObjects},
	{"📈", "chart increasing upward trend profit growth stock", grpObjects},
	{"📉", "chart decreasing downward trend loss decline", grpObjects},
	{"📊", "bar chart analytics statistics report metric", grpObjects},
	{"📋", "clipboard checklist survey audit copy paste", grpObjects},
	{"📌", "pushpin pin thumbtack note sticky location", grpObjects},
	{"📍", "round pushpin map pin location landmark marker", grpObjects},
	{"📎", "paperclip attachment attach document file", grpObjects},
	{"🖇️", "linked paperclips attachment connect link", grpObjects},
	{"📏", "straight ruler measurement length millimeter", grpObjects},
	{"📐", "triangular ruler math geometry draft architect", grpObjects},
	{"✂️", "scissors cut trim craft tailor shear", grpObjects},
	{"🗃️", "card file box archive database storage records", grpObjects},
	{"🗄️", "file cabinet filing archive drawer office", grpObjects},
	{"🗑️", "wastebasket trash bin garbage recycle delete", grpObjects},
	{"🔒", "locked padlock security private secret protect", grpObjects},
	{"🔓", "unlocked padlock open security access public", grpObjects},
	{"🔏", "locked with pen privacy signature security key", grpObjects},
	{"🔐", "locked with key password security protection safe", grpObjects},
	{"🔑", "key unlock access secret security password auth", grpObjects},
	{"🗝️", "old key vintage antique secret escape access", grpObjects},
	{"💻", "laptop computer pc portable macbook thinkpad", grpObjects},
	{"🖥️", "desktop computer monitor screen workstation pc", grpObjects},
	{"🖨️", "printer print paper office laserjet", grpObjects},
	{"⌨️", "keyboard typing input peripheral mechanical key", grpObjects},
	{"🖱️", "computer mouse track click pointer peripheral", grpObjects},
	{"🖲️", "trackball pointer mouse input device", grpObjects},
	{"💽", "minidisc optical disc storage media retro", grpObjects},
	{"💾", "floppy disk save storage 3.5 inch retro legacy", grpObjects},
	{"💿", "optical disk cd compact disc audio music album", grpObjects},
	{"📀", "dvd video digital versatile disc movie media", grpObjects},
	{"📼", "videocassette vhs tape retro film movie recording", grpObjects},
	{"📷", "camera photo photography snapshot picture image", grpObjects},
	{"📸", "camera with flash snapshot flash photo shoot", grpObjects},
	{"📹", "video camera camcorder recording movie clip", grpObjects},
	{"🎥", "movie camera cinema film filming hollywood video", grpObjects},
	{"📽️", "film projector cinema movie screening vintage", grpObjects},
	{"🎞️", "film frames photographic film cinema analog movie", grpObjects},
	{"📞", "telephone receiver call phone contact ring talk", grpObjects},
	{"☎️", "telephone vintage rotary landline phone call", grpObjects},
	{"📟", "pager beeper retro message vintage device", grpObjects},
	{"📠", "fax machine facsimile document telecopier", grpObjects},
	{"📺", "television tv screen broadcast television show", grpObjects},
	{"📻", "radio audio fm am broadcast wireless tuner", grpObjects},
	{"🎙️", "studio microphone podcast recording voice audio", grpObjects},
	{"🎚️", "level slider fader audio mixer volume equalizer", grpObjects},
	{"🎛️", "control knobs mixer dials equalizer synthesize", grpObjects},
	{"⏱️", "stopwatch timer countdown racing athletics lap", grpObjects},
	{"⏲️", "timer clock kitchen timer countdown interval alarm", grpObjects},
	{"⏰", "alarm clock wake up morning ring bell snooze", grpObjects},
	{"🕰️", "mantelpiece clock vintage antique mantle shelf", grpObjects},
	{"⌛", "hourglass done sand clock time out finished", grpObjects},
	{"⏳", "hourglass not done sand timer loading waiting", grpObjects},
	{"📱", "mobile phone smartphone iphone android cell call", grpObjects},
	{"📲", "mobile phone with arrow incoming call receive text", grpObjects},

	// ── Hearts & Emotions ─────────────────────────────────────────────────────
	{"❤️", "red heart love romance affection favorite", grpHearts},
	{"🧡", "orange heart warmth friendship care", grpHearts},
	{"💛", "yellow heart happiness joy friendship sunshine", grpHearts},
	{"💚", "green heart nature envy eco health organic", grpHearts},
	{"💙", "blue heart peace loyalty trust calm cold", grpHearts},
	{"💜", "purple heart royalty magic spirituality luxury", grpHearts},
	{"🖤", "black heart grief darkness sorrow gothic mood", grpHearts},
	{"🤍", "white heart purity clean peace innocence clear", grpHearts},
	{"🤎", "brown heart warmth earthy cozy cocoa chocolate", grpHearts},
	{"🩷", "pink heart cute sweet affection soft tenderness", grpHearts},
	{"🩵", "light blue heart sky blue calm tranquil fresh", grpHearts},
	{"🩶", "grey heart neutral silver titanium minimal subtle", grpHearts},
	{"💔", "broken heart heartbreak grief rejection break up", grpHearts},
	{"❤️‍🔥", "heart on fire burning passion intense love lust", grpHearts},
	{"❤️‍🩹", "mending heart healing recovery patch better health", grpHearts},
	{"❣️", "heart exclamation point mark love warning favor", grpHearts},
	{"💕", "two hearts pink love affection dating romance", grpHearts},
	{"💞", "revolving hearts revolving love romance revolving", grpHearts},
	{"💓", "beating heart heartbeat pulse love exciting rhythm", grpHearts},
	{"💗", "growing heart expanding love flutter affection", grpHearts},
	{"💖", "sparkling heart sparkle love shimmer glam glitter", grpHearts},
	{"💘", "heart with arrow cupid romance crush struck love", grpHearts},
	{"💝", "heart with ribbon gift present love valentine box", grpHearts},
	{"💟", "heart decoration floral ornament purple white", grpHearts},
	{"☮️", "peace symbol antiwar pacifism sign flower power", grpHearts},
	{"✝️", "latin cross christianity church religion faith", grpHearts},
	{"☪️", "star and crescent islam muslim religion symbol", grpHearts},
	{"🕉️", "om hinduism sacred sound mantra meditation yoga", grpHearts},
	{"☸️", "wheel of dharma buddhism eightfold path religion", grpHearts},
	{"✡️", "star of david judaism jewish religion hexagram", grpHearts},
	{"🔯", "six pointed star dot hexagram mystery magic sign", grpHearts},
	{"🕎", "menorah hanukkah judaism candles nine branches", grpHearts},
	{"☯️", "yin yang balance taoism harmony opposite duality", grpHearts},
	{"☦️", "orthodox cross christian orthodox church religion", grpHearts},
	{"🛐", "place of worship prayer meditation chapel shrine", grpHearts},
	{"⛎", "ophiuchus 13th zodiac sign constellation serpent", grpHearts},
	{"♈", "aries zodiac ram horoscopes astrological fire", grpHearts},
	{"♉", "taurus zodiac bull horoscopes astrological earth", grpHearts},
	{"♊", "gemini zodiac twins horoscopes astrological air", grpHearts},
	{"♋", "cancer zodiac crab horoscopes astrological water", grpHearts},
	{"♌", "leo zodiac lion horoscopes astrological fire", grpHearts},
	{"♍", "virgo zodiac maiden horoscopes astrological earth", grpHearts},
	{"♎", "libra zodiac scales horoscopes astrological air", grpHearts},
	{"♏", "scorpio zodiac scorpion horoscopes astrological water", grpHearts},
	{"♐", "sagittarius zodiac archer centaur horoscopes fire", grpHearts},
	{"♑", "capricorn zodiac sea goat horoscopes earth star", grpHearts},
	{"♒", "aquarius zodiac water bearer horoscopes air star", grpHearts},
	{"♓", "pisces zodiac fish horoscopes astrological water", grpHearts},

	// ── Nature & Weather ──────────────────────────────────────────────────────
	{"🌞", "sun with face sunshine smiling morning warmth", grpNature},
	{"🌝", "full moon with face smiling night lunar glow", grpNature},
	{"🌛", "first quarter moon face crescent smiling night", grpNature},
	{"🌜", "last quarter moon face crescent smiling sleep", grpNature},
	{"🌚", "new moon face dark moon smirk night sky", grpNature},
	{"🌕", "full moon lunar sphere night astronomy sky", grpNature},
	{"🌖", "waning gibbous moon astronomy lunar phase", grpNature},
	{"🌗", "last quarter moon half astronomy lunar phase", grpNature},
	{"🌘", "waning crescent moon astronomy lunar phase", grpNature},
	{"🌑", "new moon dark moon astronomy sky black", grpNature},
	{"🌒", "waxing crescent moon astronomy lunar phase", grpNature},
	{"🌓", "first quarter moon half astronomy lunar phase", grpNature},
	{"🌔", "waxing gibbous moon astronomy lunar phase", grpNature},
	{"🌙", "crescent moon night sky ramadan bedtime dream", grpNature},
	{"🌎", "globe showing americas earth planet world global", grpNature},
	{"🌍", "globe showing europe africa earth planet world", grpNature},
	{"🌏", "globe showing asia australia earth planet world", grpNature},
	{"🪐", "ringed planet saturn astronomy cosmos solar system", grpNature},
	{"💫", "dizzy star swirl sparkle magic comic effect", grpNature},
	{"⭐️", "white medium star astrology celestial favorite", grpNature},
	{"🌟", "glowing star shine bright sparkle award yellow", grpNature},
	{"✨", "sparkles stars clean shiny magic twinkle new", grpNature},
	{"⚡", "high voltage lightning bolt electricity power zap", grpNature},
	{"☄️", "comet meteor asteroid shooting star space cosmos", grpNature},
	{"💥", "collision explosion boom blast burst pow comic", grpNature},
	{"🔥", "fire flame burn hot blaze lit trending heat", grpNature},
	{"🌪️", "tornado cyclone twister storm whirlwind weather", grpNature},
	{"🌈", "rainbow colors pride sky weather optical arch", grpNature},
	{"☀️", "sun bright sunny clear sky weather day warmth", grpNature},
	{"🌤️", "sun behind small cloud mostly sunny fair sky", grpNature},
	{"⛅", "sun behind cloud partly cloudy overcast weather", grpNature},
	{"🌥️", "sun behind large cloud mostly cloudy day weather", grpNature},
	{"☁️", "cloud overcast cloudy sky weather computing data", grpNature},
	{"🌦️", "sun behind rain cloud sunshower precipitation", grpNature},
	{"🌧️", "cloud with rain rainy weather shower downpour storm", grpNature},
	{"🌨️", "cloud with snow snowy weather flurry winter cold", grpNature},
	{"🌩️", "cloud with lightning thunder storm thunderstorm", grpNature},
	{"❄️", "snowflake ice winter cold crystal frozen frost", grpNature},
	{"☃️", "snowman with snow winter holiday frosty festive", grpNature},
	{"⛄", "snowman winter without snow holiday frosty ice", grpNature},
	{"🌬️", "wind face blowing breeze gust weather cold air", grpNature},
	{"💨", "dashing away speed fast puff sprint escape run", grpNature},
	{"💧", "droplet water sweat tear raindrop liquid clean", grpNature},
	{"💦", "sweat droplets splash water effort splash spray", grpNature},
	{"☔", "umbrella with rain drops rainy weather shower wet", grpNature},
	{"☂️", "umbrella rain shield weather protection cover", grpNature},
	{"🌊", "water wave ocean sea surf tsunami surf tide", grpNature},
	{"🌫️", "fog mist weather hazy visibility smog cloud", grpNature},
	{"🌲", "evergreen tree pine fir forest nature woods winter", grpNature},
	{"🌳", "deciduous tree oak maple forest green wood park", grpNature},
	{"🌴", "palm tree tropical island beach coconut vacation", grpNature},
	{"🌱", "seedling sprout plant grow spring biology bot", grpNature},
	{"🌿", "herb plant mint leaf seasoning flora botanical", grpNature},
	{"☘️", "shamrock clover st patrick ireland lucky green", grpNature},
	{"🍀", "four leaf clover luck fortune lucky charm green", grpNature},
	{"🎍", "pine decoration kadomatsu japanese new year", grpNature},
	{"🪴", "potted plant houseplant indoor succulent flora", grpNature},
	{"🎋", "tanabata tree bamboo paper wishes festival japan", grpNature},
	{"🍃", "leaf fluttering in wind breeze nature foliage fall", grpNature},
	{"🍂", "fallen leaf autumn foliage dry nature brown leaf", grpNature},
	{"🍁", "maple leaf autumn fall canada foliage red orange", grpNature},
	{"🍄", "mushroom toadstool fungus mario woods forage spore", grpNature},
	{"🐚", "spiral shell seashell beach marine ocean mollusk", grpNature},
	{"🪨", "rock stone pebble boulder mineral mountain geology", grpNature},
	{"🌾", "sheaf of rice grain wheat cereal harvest agriculture", grpNature},
	{"💐", "bouquet flowers gift romance celebration wedding", grpNature},
	{"🌷", "tulip flower spring flora blossom dutch garden", grpNature},
	{"🌹", "rose flower romance love passion floral red bloom", grpNature},
	{"🥀", "wilted flower dead dying fading sad decay rose", grpNature},
	{"🌺", "hibiscus flower tropical hawaii floral exotic pink", grpNature},
	{"🌸", "cherry blossom sakura spring japan floral blossom", grpNature},
	{"🌼", "blossom yellow flower daisy floral bloom spring", grpNature},
	{"🌻", "sunflower yellow flower sunshine floral summer seed", grpNature},

	// ── Symbols & Punctuation ─────────────────────────────────────────────────
	{"✅", "check mark button green verified tick correct done", grpSymbols},
	{"❌", "cross mark x red wrong error cancel no fail", grpSymbols},
	{"❎", "cross mark button green box x error reject cancel", grpSymbols},
	{"➕", "plus sign addition positive add math calculate", grpSymbols},
	{"➖", "minus sign subtraction negative remove math deduct", grpSymbols},
	{"➗", "division sign divide math calculate ratio split", grpSymbols},
	{"✖️", "multiply sign multiplication times math calculate", grpSymbols},
	{"🟰", "heavy equals sign equality math calculate result", grpSymbols},
	{"❓", "red question mark query help punctuation confuse", grpSymbols},
	{"❔", "white question mark query help info uncertain grey", grpSymbols},
	{"❕", "white exclamation mark surprise alert notice punctuation", grpSymbols},
	{"❗", "red exclamation mark warning alert important danger", grpSymbols},
	{"‼️", "double exclamation mark shock emergency high priority", grpSymbols},
	{"⁉️", "exclamation question mark interrobang surprised doubt", grpSymbols},
	{"⚠️", "warning sign hazard alert danger caution attention", grpSymbols},
	{"🚸", "children crossing warning school pedestrian cautious", grpSymbols},
	{"⛔", "no entry stop forbidden prohibited blocked sign", grpSymbols},
	{"🚫", "prohibited sign forbidden banned restricted cancel", grpSymbols},
	{"🚳", "no bicycles no cycling prohibited bike forbidden", grpSymbols},
	{"🚭", "no smoking prohibited tobacco cigarettes forbidden", grpSymbols},
	{"🚯", "no littering prohibited trash garbage clean sign", grpSymbols},
	{"🚱", "non-potable water do not drink prohibited caution", grpSymbols},
	{"🚷", "no pedestrians prohibited walking caution keep out", grpSymbols},
	{"🔞", "no one under eighteen 18 adult mature restricted", grpSymbols},
	{"☢️", "radioactive radiation hazard nuclear danger warning", grpSymbols},
	{"☣️", "biohazard biological hazard virus danger warning", grpSymbols},
	{"🛑", "stop sign octagon halt red traffic road command", grpSymbols},
	{"🔔", "bell notification alert ring chime reminder sound", grpSymbols},
	{"🔕", "bell with slash mute silent quiet do not disturb", grpSymbols},
	{"🎵", "musical note sound melody song tune audio play", grpSymbols},
	{"🎶", "musical notes melody sound song symphony tune play", grpSymbols},
	{"💤", "zzz sleeping tired bored snooze rest nap cartoon", grpSymbols},
	{"💯", "hundred points perfect score 100 exam top excellent", grpSymbols},
	{"💢", "anger symbol vein pop anime mad annoyed furious", grpSymbols},
	{"💬", "speech balloon bubble message chat comment text talk", grpSymbols},
	{"👁️‍🗨️", "eye in speech bubble witness witness campaign speech", grpSymbols},
	{"🗯️", "right anger bubble mad shout comic scream burst", grpSymbols},
	{"💭", "thought balloon thinking daydream ponder ponder bubble", grpSymbols},
	{"💮", "white flower very well done stamp japanese grade score", grpSymbols},
	{"♨️", "hot springs onsen steam bath geothermal warm spa", grpSymbols},
	{"🌐", "globe with meridians world internet network web www", grpSymbols},
	{"💠", "diamond with a dot cute kawaii diamond gem shape", grpSymbols},
	{"Ⓜ️", "circled m metro subway underground transit logo", grpSymbols},
	{"🌀", "cyclone spiral swirl dizzy typhoon hurricane storm", grpSymbols},
	{"♿", "wheelchair symbol accessibility accessible disabled sign", grpSymbols},
	{"🅿️", "p button parking lot car garage transit sign", grpSymbols},

	// ── Arrows & Flow ─────────────────────────────────────────────────────────
	{"←", "arrow left west pointing previous back", grpArrows},
	{"→", "arrow right east pointing next forward proceed", grpArrows},
	{"↑", "arrow up north pointing upward increase lift", grpArrows},
	{"↓", "arrow down south pointing downward decrease drop", grpArrows},
	{"↔", "arrow left right horizontal width span resize bidirectional", grpArrows},
	{"↕", "arrow up down vertical height span resize bidirectional", grpArrows},
	{"↖", "arrow upper left northwest pointing top corner diagonal", grpArrows},
	{"↗", "arrow upper right northeast pointing top corner growth", grpArrows},
	{"↘", "arrow lower right southeast pointing bottom corner drop", grpArrows},
	{"↙", "arrow lower left southwest pointing bottom corner diagonal", grpArrows},
	{"↚", "leftwards arrow with stroke cancelled denied cannot", grpArrows},
	{"↛", "rightwards arrow with stroke forbidden prohibited blocked", grpArrows},
	{"↜", "leftwards wave arrow squiggly dynamic motion back", grpArrows},
	{"↝", "rightwards wave arrow squiggly dynamic motion next", grpArrows},
	{"↞", "leftwards two headed arrow fast back rewind double", grpArrows},
	{"↟", "upwards two headed arrow fast up jump top double", grpArrows},
	{"↠", "rightwards two headed arrow fast forward advance double", grpArrows},
	{"↡", "downwards two headed arrow fast down jump bottom double", grpArrows},
	{"↢", "leftwards arrow with tail origin start provenance", grpArrows},
	{"↣", "rightwards arrow with tail destination target emit", grpArrows},
	{"↤", "leftwards arrow from bar start begin point source", grpArrows},
	{"↥", "upwards arrow from bar bottom base ground lift", grpArrows},
	{"↦", "rightwards arrow from bar maps to function transformation", grpArrows},
	{"↧", "downwards arrow from bar ceiling floor drop base", grpArrows},
	{"↨", "up down arrow with base bidirectional grounded pillar", grpArrows},
	{"↩", "leftwards arrow with hook return carriage return enter line", grpArrows},
	{"↪", "rightwards arrow with hook continue branch delegate hook", grpArrows},
	{"↫", "leftwards arrow with loop loop back retry redo cycle", grpArrows},
	{"↬", "rightwards arrow with loop loop forward branch advance", grpArrows},
	{"↭", "left right wave arrow squiggly oscillation ripple", grpArrows},
	{"↮", "left right arrow with stroke cancelled bidirectional disconnected", grpArrows},
	{"↯", "downwards zigzag arrow lightning hazard breakdown shock", grpArrows},
	{"↰", "upwards arrow with tip leftwards turn angle corner route", grpArrows},
	{"↱", "upwards arrow with tip rightwards turn angle corner route", grpArrows},
	{"↲", "downwards arrow with tip leftwards turn angle corner corner", grpArrows},
	{"↳", "downwards arrow with tip rightwards sub item tree child indent", grpArrows},
	{"↵", "downwards arrow with corner leftwards enter return newline break", grpArrows},
	{"↶", "anticlockwise top semicircle arrow undo counter clockwise curve", grpArrows},
	{"↷", "clockwise top semicircle arrow redo clockwise forward curve", grpArrows},
	{"↺", "anticlockwise open circle arrow reload refresh restart undo", grpArrows},
	{"↻", "clockwise open circle arrow reload refresh rotate spin cycle", grpArrows},
	{"↼", "leftwards harpoon with barb upwards vector half arrow", grpArrows},
	{"↽", "leftwards harpoon with barb downwards vector half arrow", grpArrows},
	{"↾", "upwards harpoon with barb rightwards vector half arrow", grpArrows},
	{"↿", "upwards harpoon with barb leftwards vector half arrow", grpArrows},
	{"⇀", "rightwards harpoon with barb upwards vector half arrow chemistry", grpArrows},
	{"⇁", "rightwards harpoon with barb downwards vector half arrow chemistry", grpArrows},
	{"⇂", "downwards harpoon with barb rightwards vector half arrow", grpArrows},
	{"⇃", "downwards harpoon with barb leftwards vector half arrow", grpArrows},
	{"⇄", "rightwards arrow over leftwards arrow exchange swap trade flow", grpArrows},
	{"⇅", "upwards arrow leftwards of downwards arrow duplex vertical flow", grpArrows},
	{"⇆", "leftwards arrow over rightwards arrow counter trade reverse flow", grpArrows},
	{"⇇", "leftwards paired arrows parallel double back left stream", grpArrows},
	{"⇈", "upwards paired arrows parallel double rise up stream", grpArrows},
	{"⇉", "rightwards paired arrows parallel double forward advance stream", grpArrows},
	{"⇊", "downwards paired arrows parallel double drop sink stream", grpArrows},
	{"⇋", "leftwards harpoon over rightwards harpoon equilibrium reaction reversible", grpArrows},
	{"⇌", "rightwards harpoon over leftwards harpoon chemical equilibrium reversible", grpArrows},
	{"⇐", "leftwards double arrow implies from logic back condition", grpArrows},
	{"⇑", "upwards double arrow lift double strong up rise", grpArrows},
	{"⇒", "rightwards double arrow implies then logic forward condition", grpArrows},
	{"⇓", "downwards double arrow drop double strong down sink", grpArrows},
	{"⇔", "left right double arrow iff equivalent biconditional logic", grpArrows},
	{"⇕", "up down double arrow vertical equivalence span logic", grpArrows},
	{"⇖", "north west double arrow diagonal double top left", grpArrows},
	{"⇗", "north east double arrow diagonal double top right", grpArrows},
	{"⇘", "south east double arrow diagonal double bottom right", grpArrows},
	{"⇙", "south west double arrow diagonal double bottom left", grpArrows},
	{"⇚", "leftwards triple arrow very fast rewind triple power back", grpArrows},
	{"⇛", "rightwards triple arrow very fast forward triple power advance", grpArrows},
	{"⇜", "leftwards squiggly arrow irregular motion zigzag back", grpArrows},
	{"⇝", "rightwards squiggly arrow irregular motion zigzag next", grpArrows},
	{"⇞", "upwards arrow with double stroke page up fast leap", grpArrows},
	{"⇟", "downwards arrow with double stroke page down fast drop", grpArrows},
	{"⇠", "leftwards dashed arrow dotted dashed trail back", grpArrows},
	{"⇡", "upwards dashed arrow dotted dashed trail rise", grpArrows},
	{"⇢", "rightwards dashed arrow dotted dashed trail advance", grpArrows},
	{"⇣", "downwards dashed arrow dotted dashed trail descend", grpArrows},
	{"⇤", "leftwards arrow to bar tab previous start limit stop", grpArrows},
	{"⇥", "rightwards arrow to bar tab next stop limit end", grpArrows},
	{"⇦", "leftwards white arrow hollow bold directional button left", grpArrows},
	{"⇨", "rightwards white arrow hollow bold directional button right", grpArrows},
	{"⇧", "upwards white arrow hollow shift caps uppercase button", grpArrows},
	{"⇩", "downwards white arrow hollow bold directional button down", grpArrows},
	{"➔", "heavy rightwards arrow thick bold bullet forward indicator", grpArrows},
	{"⟵", "long leftwards arrow wide extended span return flow", grpArrows},
	{"⟶", "long rightwards arrow wide extended span forward arrow", grpArrows},
	{"⟷", "long left right arrow wide extended bidirectional span", grpArrows},
	{"⟹", "long rightwards double arrow extended implication consequence", grpArrows},
	{"⟺", "long left right double arrow extended equivalence logic", grpArrows},

	// ── Block elements & Shapes ───────────────────────────────────────────────
	{"█", "full block solid black rectangle filled cell", grpBlocks},
	{"▓", "dark shade high density textured pattern fill 75 percent", grpBlocks},
	{"▒", "medium shade medium density textured pattern fill 50 percent", grpBlocks},
	{"░", "light shade low density textured pattern fill 25 percent", grpBlocks},
	{"▀", "upper half block top solid half cell ceiling", grpBlocks},
	{" ", "lower one eighth block bottom floor slice 1/8", grpBlocks},
	{"▂", "lower one quarter block bottom floor slice 1/4", grpBlocks},
	{"▃", "lower three eighths block bottom floor slice 3/8", grpBlocks},
	{"▄", "lower half block bottom solid half cell floor 1/2", grpBlocks},
	{"▅", "lower five eighths block bottom floor slice 5/8", grpBlocks},
	{"▆", "lower three quarters block bottom floor slice 3/4", grpBlocks},
	{"▇", "lower seven eighths block bottom floor slice 7/8", grpBlocks},
	{"▉", "left seven eighths block left pillar bar 7/8", grpBlocks},
	{"▊", "left three quarters block left pillar bar 3/4", grpBlocks},
	{"▋", "left five eighths block left pillar bar 5/8", grpBlocks},
	{"▌", "left half block left solid half cell pillar 1/2", grpBlocks},
	{"▍", "left three eighths block left pillar bar 3/8", grpBlocks},
	{"▎", "left one quarter block left pillar bar 1/4", grpBlocks},
	{"▏", "left one eighth block left thin edge line 1/8", grpBlocks},
	{"▐", "right half block right solid half cell wall 1/2", grpBlocks},
	{"▔", "upper one eighth block top thin roof edge ceiling 1/8", grpBlocks},
	{"▕", "right one eighth block right thin wall edge 1/8", grpBlocks},
	{"▖", "quadrant lower left corner block pixel", grpBlocks},
	{"▗", "quadrant lower right corner block pixel", grpBlocks},
	{"▘", "quadrant upper left corner block pixel", grpBlocks},
	{"▙", "quadrant upper left lower left lower right three quarters", grpBlocks},
	{"▚", "quadrant upper left lower right checker diagonal", grpBlocks},
	{"▛", "quadrant upper left upper right lower left three quarters", grpBlocks},
	{"▜", "quadrant upper left upper right lower right three quarters", grpBlocks},
	{"▝", "quadrant upper right corner block pixel", grpBlocks},
	{"▞", "quadrant upper right lower left checker diagonal", grpBlocks},
	{"▟", "quadrant upper right lower left lower right three quarters", grpBlocks},
	{"■", "black square solid filled dark geometric box", grpBlocks},
	{"□", "white square hollow outline light geometric box", grpBlocks},
	{"▢", "white square rounded corners soft outline frame", grpBlocks},
	{"▣", "white square containing black small square checkbox select", grpBlocks},
	{"▤", "square with horizontal fill lined striped pattern", grpBlocks},
	{"▥", "square with vertical fill striped barcode pattern", grpBlocks},
	{"▦", "square with orthogonal crosshatch grid mesh pattern", grpBlocks},
	{"▧", "square with upper left to lower right fill diagonal hatch", grpBlocks},
	{"▨", "square with upper right to lower left fill diagonal hatch", grpBlocks},
	{"▩", "square with diagonal crosshatch mesh diamond pattern", grpBlocks},
	{"▪", "black small square bullet point dark tiny box", grpBlocks},
	{"▫", "white small square bullet point hollow tiny box", grpBlocks},
	{"▬", "black rectangle solid horizontal bar dark beam", grpBlocks},
	{"▭", "white rectangle hollow horizontal bar frame beam", grpBlocks},
	{"▮", "black vertical rectangle solid vertical pillar dark", grpBlocks},
	{"▯", "white vertical rectangle hollow vertical pillar frame", grpBlocks},
	{"▰", "black parallelogram solid slanted quadrilateral", grpBlocks},
	{"▱", "white parallelogram hollow slanted quadrilateral", grpBlocks},
	{"▲", "black up pointing triangle solid filled arrow top", grpBlocks},
	{"△", "white up pointing triangle hollow outline arrow top", grpBlocks},
	{"▴", "black up pointing small triangle tiny solid arrow", grpBlocks},
	{"▵", "white up pointing small triangle tiny hollow arrow", grpBlocks},
	{"▶", "black right pointing triangle solid play advance next", grpBlocks},
	{"▷", "white right pointing triangle hollow play advance next", grpBlocks},
	{"▸", "black right pointing small triangle tiny solid next", grpBlocks},
	{"▹", "white right pointing small triangle tiny hollow next", grpBlocks},
	{"►", "black right pointing pointer solid forward pointer", grpBlocks},
	{"▼", "black down pointing triangle solid filled arrow drop", grpBlocks},
	{"▽", "white down pointing triangle hollow outline arrow drop", grpBlocks},
	{"▾", "black down pointing small triangle tiny solid drop", grpBlocks},
	{"▿", "white down pointing small triangle tiny hollow drop", grpBlocks},
	{"◀", "black left pointing triangle solid back previous", grpBlocks},
	{"◁", "white left pointing triangle hollow back previous", grpBlocks},
	{"◂", "black left pointing small triangle tiny solid back", grpBlocks},
	{"◃", "white left pointing small triangle tiny hollow back", grpBlocks},
	{"◄", "black left pointing pointer solid back pointer", grpBlocks},
	{"◆", "black diamond solid filled rhombus gem dark", grpBlocks},
	{"◇", "white diamond hollow outline rhombus gem frame", grpBlocks},
	{"◈", "white diamond containing black small diamond target", grpBlocks},
	{"◉", "fisheye circle radio button target bullseye selected", grpBlocks},
	{"◊", "lozenge diamond outline shape bullet math", grpBlocks},
	{"○", "white circle hollow ring circular outline unselected", grpBlocks},
	{"◌", "dotted circle placeholder anchor dashed ring", grpBlocks},
	{"◍", "circle with vertical fill striped sphere texture", grpBlocks},
	{"◎", "bullseye target concentric circles outer inner", grpBlocks},
	{"●", "black circle solid filled dot disk bullet point", grpBlocks},
	{"◐", "circle with left half black half filled moon phase", grpBlocks},
	{"◑", "circle with right half black half filled moon phase", grpBlocks},
	{"◒", "circle with lower half black half filled sphere bottom", grpBlocks},
	{"◓", "circle with upper half black half filled sphere top", grpBlocks},
	{"◔", "circle with upper right quadrant black pie chart 25 percent", grpBlocks},
	{"◕", "circle with all but upper left quadrant black pie chart 75", grpBlocks},
	{"◖", "left half black circle solid semi circle crescent", grpBlocks},
	{"◗", "right half black circle solid semi circle crescent", grpBlocks},
	{"◘", "inverse bullet square containing white circle target", grpBlocks},
	{"◙", "inverse white circle containing black small square ring", grpBlocks},
	{"◢", "black lower right triangle solid wedge slope corner", grpBlocks},
	{"◣", "black lower left triangle solid wedge slope corner", grpBlocks},
	{"◤", "black upper left triangle solid wedge slope corner", grpBlocks},
	{"◥", "black upper right triangle solid wedge slope corner", grpBlocks},
	{"◯", "large circle hollow outline big ring hoop zero", grpBlocks},

	// ── Line drawing ──────────────────────────────────────────────────────────
	{"─", "box light horizontal single line rule border dash", grpLines},
	{"━", "box heavy horizontal bold thick line rule beam", grpLines},
	{"│", "box light vertical single line rule border pipe", grpLines},
	{"┃", "box heavy vertical bold thick line rule column", grpLines},
	{"┄", "box light triple dash horizontal dashed line rule", grpLines},
	{"┅", "box heavy triple dash horizontal bold dashed rule", grpLines},
	{"┆", "box light triple dash vertical dashed line column", grpLines},
	{"┇", "box heavy triple dash vertical bold dashed column", grpLines},
	{"┈", "box light quadruple dash horizontal dotted dash line", grpLines},
	{"┉", "box heavy quadruple dash horizontal bold dotted dash", grpLines},
	{"┊", "box light quadruple dash vertical dotted line column", grpLines},
	{"┋", "box heavy quadruple dash vertical bold dotted column", grpLines},
	{"═", "box double horizontal double rule line parallel beam", grpLines},
	{"║", "box double vertical double rule line parallel column", grpLines},
	{"╌", "box light double dash horizontal dotted line", grpLines},
	{"╍", "box heavy double dash horizontal bold dotted line", grpLines},
	{"╎", "box light double dash vertical dotted column", grpLines},
	{"╏", "box heavy double dash vertical bold dotted column", grpLines},
	{"╴", "box light left single horizontal line cap terminal", grpLines},
	{"╵", "box light up single vertical line cap bottom terminal", grpLines},
	{"╶", "box light right single horizontal line cap start terminal", grpLines},
	{"╷", "box light down single vertical line cap top terminal", grpLines},
	{"╸", "box heavy left thick horizontal line cap end", grpLines},
	{"╹", "box heavy up thick vertical line cap bottom", grpLines},
	{"╺", "box heavy right thick horizontal line cap start", grpLines},
	{"╻", "box heavy down thick vertical line cap top", grpLines},
	{"╼", "box light left heavy right horizontal transition", grpBox},
	{"╽", "box light up heavy down vertical transition", grpBox},
	{"╾", "box heavy left light right horizontal transition", grpBox},
	{"╿", "box heavy up light down vertical transition", grpBox},
	{"╒", "box down single right double corner top left mixed", grpBox},
	{"╓", "box down double right single corner top left mixed", grpBox},
	{"╔", "box double down double right corner top left double", grpBox},
	{"╕", "box down single left double corner top right mixed", grpBox},
	{"╖", "box down double left single corner top right mixed", grpBox},
	{"╗", "box double down double left corner top right double", grpBox},
	{"╘", "box up single right double corner bottom left mixed", grpBox},
	{"╙", "box up double right single corner bottom left mixed", grpBox},
	{"╚", "box double up double right corner bottom left double", grpBox},
	{"╛", "box up single left double corner bottom right mixed", grpBox},
	{"╜", "box up double left single corner bottom right mixed", grpBox},
	{"╝", "box double up double left corner bottom right double", grpBox},
	{"╞", "box vertical single right double tee left junction mixed", grpBox},
	{"╟", "box vertical double right single tee left junction mixed", grpBox},
	{"╠", "box double vertical double right tee left junction double", grpBox},
	{"╡", "box vertical single left double tee right junction mixed", grpBox},
	{"╢", "box vertical double left single tee right junction mixed", grpBox},
	{"╣", "box double vertical double left tee right junction double", grpBox},
	{"╤", "box down single horizontal double tee top junction mixed", grpBox},
	{"╥", "box down double horizontal single tee top junction mixed", grpBox},
	{"╦", "box double down double horizontal tee top junction double", grpBox},
	{"╧", "box up single horizontal double tee bottom junction mixed", grpBox},
	{"╨", "box up double horizontal single tee bottom junction mixed", grpBox},
	{"╩", "box double up double horizontal tee bottom junction double", grpBox},
	{"╪", "box vertical single horizontal double cross four way mixed", grpBox},
	{"╫", "box vertical double horizontal single cross four way mixed", grpBox},
	{"╬", "box double vertical double horizontal cross four way double", grpBox},

	// ── Box-drawing ───────────────────────────────────────────────────────────
	{"┌", "box light down and right sharp top left corner corner", grpBox},
	{"┐", "box light down and left sharp top right corner corner", grpBox},
	{"└", "box light up and right sharp bottom left corner corner", grpBox},
	{"┘", "box light up and left sharp bottom right corner corner", grpBox},
	{"├", "box light vertical and right tee left junction branch", grpBox},
	{"┤", "box light vertical and left tee right junction branch", grpBox},
	{"┬", "box light down and horizontal tee top junction roof", grpBox},
	{"┴", "box light up and horizontal tee bottom junction base", grpBox},
	{"┼", "box light vertical and horizontal cross four way intersection plus", grpBox},
	{"┏", "box heavy down and right bold thick top left corner", grpBox},
	{"┓", "box heavy down and left bold thick top right corner", grpBox},
	{"┗", "box heavy up and right bold thick bottom left corner", grpBox},
	{"┛", "box heavy up and left bold thick bottom right corner", grpBox},
	{"┣", "box heavy vertical and right bold thick tee left junction", grpBox},
	{"┫", "box heavy vertical and left bold thick tee right junction", grpBox},
	{"┳", "box heavy down and horizontal bold thick tee top junction", grpBox},
	{"┻", "box heavy up and horizontal bold thick tee bottom junction", grpBox},
	{"╋", "box heavy vertical and horizontal bold thick cross intersection", grpBox},
	{"╭", "box light arc down and right rounded soft top left corner", grpBox},
	{"╮", "box light arc down and left rounded soft top right corner", grpBox},
	{"╯", "box light arc up and left rounded soft bottom right corner", grpBox},
	{"╰", "box light arc up and right rounded soft bottom left corner", grpBox},
	{"╱", "box light diagonal upper right to lower left slash forward", grpBox},
	{"╲", "box light diagonal upper left to lower right backslash backward", grpBox},
	{"╳", "box light diagonal cross diagonal intersection cross mark", grpBox},
}

const (
	emojiGridColumns = 10
)

// ── Widget types ──────────────────────────────────────────────────────────────

type picker struct {
	query      *loom.TextInput
	split      *loom.Split
	group      int
	items      []int
	index      int
	gridFocus  bool // true = grid is focused (arrow keys move selection); false = search is focused
	chosen     string
	width      int
	searchY    int
	categoryY  int
	lastRect   loom.Rect
	gridRect   loom.Rect
	cols       int
	viewRows   int
	gridStart  int
	categories []category
	entries    []entry
	cellWidth  int
}

type gridPane struct{ picker *picker }

func (g *gridPane) Draw(c *loom.Canvas, r loom.Rect)          { g.picker.drawGrid(c, r) }
func (g *gridPane) ConsumeKey(loom.KeyEvent) loom.EventResult { return loom.Ignored() }
func (g *gridPane) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if g.picker.handleGridMouse(e) {
		return loom.QuitResult()
	}
	return loom.Ignored()
}

func newPicker() *picker {
	p := &picker{
		query:      loom.NewTextInput(""),
		gridFocus:  true, // start with grid focused so arrow keys work immediately
		categories: categories(),
		entries:    entries(),
	}
	p.split = loom.NewSplit(&gridPane{picker: p}, nil)
	p.split.Ratio = 1.0
	p.split.MinFirst = 10
	p.split.MinSecond = 0
	p.refresh()
	return p
}

func (p *picker) refresh() {
	p.items = p.items[:0]
	q := strings.TrimSpace(strings.ToLower(p.query.Value()))
	for i, e := range p.entries {
		matchQuery := q == "" || strings.Contains(strings.ToLower(e.name), q) || strings.Contains(e.icon, q)
		if q == "" {
			if e.group == p.group {
				p.items = append(p.items, i)
			}
		} else if matchQuery {
			p.items = append(p.items, i)
		}
	}
	p.index = min(p.index, max(0, len(p.items)-1))
}

func (p *picker) selectCategory(idx int) {
	if idx >= 0 && idx < len(p.categories) {
		p.group = idx
		p.query.SetValue("")
		p.index = 0
		p.gridFocus = true
		p.refresh()
	}
}

func (p *picker) prevCategory() {
	p.selectCategory((p.group - 1 + len(p.categories)) % len(p.categories))
}

func (p *picker) nextCategory() {
	p.selectCategory((p.group + 1) % len(p.categories))
}

// ── Draw ──────────────────────────────────────────────────────────────────────

func (p *picker) Draw(c *loom.Canvas, r loom.Rect) {
	p.lastRect = r
	if r.W <= 0 || r.H <= 0 {
		return
	}
	background := loom.ColorRGB(38, 40, 43)
	panel := loom.ColorRGB(31, 33, 36)
	accent := loom.ColorRGB(32, 151, 185)

	c.PaintSurface(r, loom.Style{BG: background})
	// Left border strip
	for y := r.Y; y < r.Y+r.H; y++ {
		c.PaintForeground(r.X, y, loom.Cell{Text: " ", Style: loom.Style{BG: panel}, Claim: true})
	}
	// Top border strip
	for x := r.X; x < r.X+r.W; x++ {
		c.PaintForeground(x, r.Y, loom.Cell{Text: " ", Style: loom.Style{BG: panel}, Claim: true})
	}

	p.width = r.W

	// ── Search bar ────────────────────────────────────────────────────────────
	p.searchY = r.Y + 1
	searchRect := loom.Rect{X: r.X, Y: p.searchY, W: r.W, H: min(2, max(0, r.Y+r.H-p.searchY))}
	c.PaintSurface(searchRect, loom.Style{BG: panel})
	if r.H >= 3 {
		c.PaintSurface(loom.Rect{X: r.X + 1, Y: p.searchY, W: max(0, r.W-2), H: 1}, loom.Style{BG: loom.ColorRGB(67, 69, 72)})
		c.Write(r.X+2, p.searchY, "⌕", loom.Style{FG: accent, Bold: true})
		query := p.query.Value()
		if query == "" {
			if p.gridFocus {
				c.Write(r.X+5, p.searchY, "search…", loom.Style{FG: loom.ColorRGB(120, 120, 120), Dim: true})
			} else {
				c.Write(r.X+5, p.searchY, "search…", loom.Style{FG: loom.ColorRGB(180, 180, 180)})
			}
		} else {
			c.Write(r.X+5, p.searchY, query, loom.Style{FG: loom.ColorRGB(220, 220, 220)})
		}
		if !p.gridFocus {
			// Show cursor in search bar when search is focused
			caret := min(p.query.Caret(), len([]rune(p.query.Value())))
			c.CursorX = r.X + 5 + loom.StringWidth(string([]rune(p.query.Value())[:caret]))
			c.CursorY = p.searchY
		}
	}

	// ── Category bar ──────────────────────────────────────────────────────────
	p.categoryY = r.Y + r.H - 2
	if p.categoryY >= r.Y && p.categoryY < r.Y+r.H {
		c.PaintSurface(loom.Rect{X: r.X, Y: p.categoryY, W: r.W, H: 1}, loom.Style{BG: panel})
		x := r.X + 2
		for i, cat := range p.categories {
			w := loom.StringWidth(cat.icon)
			if x+w >= r.X+r.W-1 {
				break
			}
			style := loom.Style{FG: loom.ColorRGB(190, 190, 190)}
			if i == p.group {
				style = loom.Style{FG: loom.ColorRGB(255, 255, 255), BG: accent, Bold: true}
			}
			c.Write(x, p.categoryY, cat.icon, style)
			x += w + 2
		}
	}

	// ── Status bar ────────────────────────────────────────────────────────────
	statusY := r.Y + r.H - 1
	c.PaintSurface(loom.Rect{X: r.X, Y: statusY, W: r.W, H: 1}, loom.Style{BG: panel})
	footer := "↑↓←→ grid   Tab search   f/[ ] cycle   Enter copy   F10 Quit"
	if len(p.items) > 0 {
		e := p.entries[p.items[p.index]]
		cat := p.categories[e.group].label
		footer = fmt.Sprintf("%s  %s  [%s]  %d / %d", e.icon, e.name, cat, p.index+1, len(p.items))
	}
	c.Write(r.X+1, statusY, footer, loom.Style{FG: loom.ColorRGB(175, 178, 181), Dim: true})

	// ── Grid area ─────────────────────────────────────────────────────────────
	mainY := p.searchY + 2
	mainRect := loom.Rect{X: r.X, Y: mainY, W: r.W, H: max(0, p.categoryY-mainY)}
	p.gridRect = mainRect
	p.split.Draw(c, mainRect)
}

func (p *picker) drawGrid(c *loom.Canvas, r loom.Rect) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	accent := loom.ColorRGB(32, 151, 185)
	p.cellWidth = 1
	for _, item := range p.items {
		icon := p.entries[item].icon
		p.cellWidth = max(p.cellWidth, loom.StringWidth(icon))
	}
	p.cellWidth = max(1, p.cellWidth)
	columnWidth := p.cellWidth + 1
	p.cols = min(emojiGridColumns, max(1, (r.W-2)/columnWidth))
	p.viewRows = r.H
	selectedRow := p.index / p.cols
	startRow := 0
	if selectedRow >= p.viewRows && p.viewRows > 0 {
		startRow = selectedRow - p.viewRows + 1
	}
	start := startRow * p.cols
	p.gridStart = start
	visible := min(len(p.items)-start, p.cols*p.viewRows)
	for n := 0; n < visible; n++ {
		idx := start + n
		x := 1 + (n%p.cols)*columnWidth
		y := n / p.cols
		icon := p.entries[p.items[idx]].icon
		style := loom.Style{FG: loom.ColorRGB(220, 200, 120)}
		cellText := loom.TruncateText(icon, p.cellWidth, "")
		if idx == p.index {
			style = loom.Style{FG: loom.ColorRGB(255, 255, 255), BG: accent, Bold: true}
			if w := loom.StringWidth(cellText); w < p.cellWidth {
				cellText += strings.Repeat(" ", p.cellWidth-w)
			}
		}
		c.Write(r.X+x, r.Y+y, cellText, style)
	}
	if len(p.items) == 0 {
		c.Write(r.X+2, r.Y, "No matching entries", loom.Style{FG: loom.ColorRGB(170, 170, 170), Dim: true})
	}
}

// ── Input handling ────────────────────────────────────────────────────────────

func (p *picker) ConsumeKey(e loom.KeyEvent) loom.EventResult {
	key := e.Key
	if key == "" {
		key = e.Text
	}
	switch key {
	case "ctrl-c", "esc":
		return loom.QuitResult()
	case "tab":
		p.gridFocus = !p.gridFocus
	case "left":
		if !p.gridFocus {
			p.query.ConsumeKey(e)
			p.refresh()
		} else if p.index > 0 {
			p.index--
		}
	case "right":
		if !p.gridFocus {
			p.query.ConsumeKey(e)
			p.refresh()
		} else if p.index+1 < len(p.items) {
			p.index++
		}
	case "up":
		p.gridFocus = true
		p.index = max(0, p.index-p.cols)
	case "down":
		p.gridFocus = true
		p.index = min(max(0, len(p.items)-1), p.index+p.cols)
	case "pgup", "shift-left":
		p.prevCategory()
	case "pgdown", "shift-right", "ctrl-f":
		p.nextCategory()
	case "enter":
		if len(p.items) > 0 {
			p.chosen = p.entries[p.items[p.index]].icon
			return loom.QuitResult()
		}
	case "backspace":
		if !p.gridFocus {
			p.query.ConsumeKey(e)
			p.refresh()
		}
	default:
		// [ / ] or f / F cycle categories when in grid mode.
		if p.gridFocus && (key == "[" || key == "]" || key == "f" || key == "F") {
			if key == "[" {
				p.prevCategory()
			} else {
				p.nextCategory()
			}
			return loom.Ignored()
		}
		// Printable text types into search and switches to search mode if in grid.
		if e.Text != "" {
			p.gridFocus = false
			p.query.ConsumeKey(e)
			p.index = 0
			p.refresh()
		}
	}
	return loom.Ignored()
}

func (p *picker) ConsumeMouse(e loom.MouseEvent) loom.EventResult {
	if e.Action != loom.MousePress && e.Action != loom.MouseHover && e.Action != loom.MouseDrag {
		return loom.Ignored()
	}
	x, y := e.X+p.lastRect.X, e.Y+p.lastRect.Y
	// Click on search bar → switch to search focus
	if (y == p.searchY || y == p.searchY+1) && e.Action == loom.MousePress && e.Button == loom.MouseLeft {
		p.gridFocus = false
		return loom.Ignored()
	}
	// Click on category bar → switch category
	if y == p.categoryY || y == p.categoryY+1 {
		if e.Action != loom.MousePress || e.Button != loom.MouseLeft {
			return loom.Ignored()
		}
		if x < p.lastRect.X+2 || x >= p.lastRect.X+p.width {
			return loom.Ignored()
		}
		// Walk the category icons to find which one was clicked.
		cx := p.lastRect.X + 2
		for i, cat := range p.categories {
			w := loom.StringWidth(cat.icon) + 2
			if x >= cx && x < cx+w {
				p.selectCategory(i)
				return loom.Ignored()
			}
			cx += w
		}
		return loom.Ignored()
	}
	e.X, e.Y = x-p.gridRect.X, y-p.gridRect.Y
	return p.split.ConsumeMouse(e)
}

func (p *picker) handleGridMouse(e loom.MouseEvent) bool {
	if e.Action != loom.MousePress && e.Action != loom.MouseHover && e.Action != loom.MouseDrag {
		return false
	}
	if e.Y < 0 || e.Y >= p.viewRows || e.X < 1 {
		return false
	}
	if p.cellWidth < 1 {
		p.cellWidth = 1
	}
	col := (e.X - 1) / max(1, p.cellWidth+1)
	idx := p.gridStart + e.Y*p.cols + col
	if col < 0 || col >= p.cols || idx < 0 || idx >= len(p.items) {
		return false
	}
	p.index, p.gridFocus = idx, true
	if e.Action == loom.MousePress && e.Button == loom.MouseLeft {
		p.chosen = p.entries[p.items[idx]].icon
		return true
	}
	return false
}

// ── Public API ────────────────────────────────────────────────────────────────

// NewWidget constructs the picker widget for embedding in another Loom app.
func NewWidget(args []string) (loom.Widget, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("loomoji: unexpected arguments: %v", args)
	}
	return newPicker(), nil
}

// Run launches the interactive emoji and symbol picker.
func Run(args []string) error {
	cmd := newCommand(runPicker, runMeasure)
	cmd.SetArgs(args)
	return cmd.Execute()
}

func runPicker() error {
	pane, err := loom.New(13)
	if err != nil {
		return err
	}
	defer pane.Close()
	pane.Resizeable = true
	pane.EnableMouse()
	app := newPicker()
	if err := pane.Run(app); err != nil {
		return err
	}
	if app.chosen != "" {
		_, err = fmt.Fprint(os.Stdout, app.chosen)
		return err
	}
	return nil
}
