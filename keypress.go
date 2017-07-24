// keypress reads a list of keystrokes from standard input
// to emulate
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"unsafe"

	"github.com/as/cursor"
)

var (
	Down = Key{
		kind:  KindKeyboard,
		flags: 0,
	}
	Up = Key{
		kind:  KindKeyboard,
		flags: FlagUp,
	}
)
var (
	Input  Key
	M      Mouse
	InputL = int(unsafe.Sizeof(Input))
)

type Flag int32

type Key struct {
	kind  int32   // 4
	_     uint32  // 4
	vk    VK      // 4
	scan  int16   // 4
	flags Flag    // 4
	time  int32   // 4
	extra uintptr // 8
	_     [2]byte
}
type Mouse struct {
	kind   int32
	x      int32 // 4
	y      int32 // 4	// good
	data   int32 // 4
	flag   int32 // 4
	time   int32 // 4
	extra1 int64 // 8
	r1     [8]byte
}

const (
	HookKbd      = 2
	HookKbdLow   = 13
	HookMouse    = 7
	HookMouseLow = 14
)

const (
	KindKeyboard = 1
	KindMouse    = 0
	KindDevice   = 2
)
const (
	FlagExtended Flag = 1 // scan code  preceded by prefix byte 0xE0 (224).
	FlagUp            = 2 // key released. If not specified, the key pressed.
	FlagScanCode      = 8 // scan identifies; wVk ignored.
	FlagUnicode       = 4 //
)
const (
	MouseAbs   = 0x8000 // Packers dont use absolute coordinates by default
	MouseMoved = 0x0001 // A movement occured, but was it _desired_?
)

func init() {
	println(unsafe.Sizeof(M))
}

//sys	SendInput(nin int, in uintptr, inlen int) (err error) = user32.SendInput
func Send(in []Key) (err error) {
	err = SendInput(
		len(in),
		uintptr(unsafe.Pointer(&(in[0]))),
		InputL,
	)
	return err
}
func SendMouse(in []Mouse) (err error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	x := in[0]
	binary.Write(buf, binary.LittleEndian, x.kind)
	binary.Write(buf, binary.LittleEndian, x.kind)
	binary.Write(buf, binary.LittleEndian, x.x)
	binary.Write(buf, binary.LittleEndian, x.y)
	binary.Write(buf, binary.LittleEndian, x.data)
	binary.Write(buf, binary.LittleEndian, x.flag)
	binary.Write(buf, binary.LittleEndian, x.time)
	binary.Write(buf, binary.LittleEndian, x.extra1)
	binary.Write(buf, binary.LittleEndian, []byte("\x00\x00\x00\x00"))
	log.Println(len(buf.Bytes()))
	err = SendInput(
		len(in),
		uintptr(unsafe.Pointer(&(buf.Bytes()[0]))),
		40,
	)
	return err
}

var (
	I = flag.Duration("i", 150*time.Millisecond, "one-time initialization delay")
	m = flag.Bool("m", false, "mouse")
	r = flag.Bool("r", false, "relative coordinates")
	w = flag.Bool("w", false, "write")
)

func init() {
	flag.Parse()
}

func mouse() {
	time.Sleep(*I)
	for sc := bufio.NewScanner(os.Stdin); sc.Scan(); {
		ln := len(sc.Bytes())
		if ln == 0 {
			continue
		}
		if ln != 49 {
			println(ln)
			continue
		}
		x, y, btn, ms, err := cursor.ReadString(sc.Text())
		if err != nil {
			log.Println(err)
		}
		log.Println(cursor.ReadString(sc.Text()))
		btn = btn
		fl := int32(btn)
		fl = 0
		m := int32(1)
		if btn == 1 && !*r {
			fl = 0x8000
			m *= 10
		}
		println(btn)
		err = SendMouse([]Mouse{
			Mouse{
				kind: KindMouse,
				x:    int32(x) * m,
				y:    int32(y) * m,
				flag: fl | int32(btn),
				time: int32(ms),
			},
		})
		if err != nil {
			log.Printf("Send: %s\n", err)
		}
	}
}
var cb = func(ncode int, wp int64, lp int) uintptr {

				fd, _ := os.Create(`C:\users\as\proc.txt`)
				fmt.Fprintln(fd,ncode, wp, lp)
				log.Println( ncode, wp, lp)
				fd.Close()
				syscall.Syscall(404, 1, 0, 0, 0)
				return 0
			}
func main() {
	if *w {
		runtime.LockOSThread()
		log.Println(SetWindowsHookEx(HookMouseLow, syscall.NewCallbackCDecl(cb), 0, 0))
		for {
		}
		return
	}
	if *m {
		mouse()
	} else {
		key()
	}
}
func key() {
	panic("no")
	buf := make([]Key, 0, 128)
	time.Sleep(*I)
	for sc := bufio.NewScanner(os.Stdin); sc.Scan(); {
		ln := sc.Bytes()
		if len(ln) == 0 {
			continue
		}
		toRelease := make([]Key, 0, 4)
		for i := 0; i < len(ln); i++ {
			switch ln[i] {
			case '#', '!', '^', '+':
				key := Down
				key.vk = special[ln[i]]
				buf = append(buf, key)
				key.flags = FlagUp
				toRelease = append([]Key{key}, toRelease...)
			default:
				if ln[i] == '{' {
					expr, adv, err := parseBraceExpr(ln[i:])
					if err != nil {
						log.Printf("parseBraceExpr: %s\n", err)
					}
					keys, err := Lookup(expr)
					if err != nil {
						log.Printf("Lookup: %s\n", err)
					}
					buf = append(buf, keys...)
					i += adv
				} else {
					buf = append(buf, Translate(ln[i])...)
				}
				buf = append(buf, toRelease...)
				toRelease = toRelease[:0]
			}
		}
		err := Send(buf)
		if err != nil {
			log.Printf("Send: %s\n", err)
		}
		buf = buf[:0]
	}
}

func Lookup(expr BraceExpr) (keys []Key, err error) {
	up := Up
	dn := Down
	vk, ok := keynames[expr.key]
	if !ok {
		return nil, fmt.Errorf("cant find key for expr %#v", expr)
	}
	up.vk = vk
	dn.vk = vk
	return []Key{up, dn}, nil
}

var special = map[byte]VK{
	'#': VK_LWIN,
	'!': VK_RMENU,
	'^': VK_CONTROL,
	'+': VK_RSHIFT,
}

var keynames = map[string]VK{
	"LBUTTON":             VK_LBUTTON,
	"RBUTTON":             VK_RBUTTON,
	"CANCEL":              VK_CANCEL,
	"MBUTTON":             VK_MBUTTON,
	"XBUTTON1":            VK_XBUTTON1,
	"XBUTTON2":            VK_XBUTTON2,
	"BACK":                VK_BACK,
	"TAB":                 VK_TAB,
	"CLEAR":               VK_CLEAR,
	"RETURN":              VK_RETURN,
	"SHIFT":               VK_SHIFT,
	"CONTROL":             VK_CONTROL,
	"MENU":                VK_MENU,
	"PAUSE":               VK_PAUSE,
	"CAPITAL":             VK_CAPITAL,
	"KANA":                VK_KANA,
	"HANGUEL":             VK_HANGUEL,
	"HANGUL":              VK_HANGUL,
	"JUNJA":               VK_JUNJA,
	"FINAL":               VK_FINAL,
	"KANJI":               VK_KANJI,
	"ESCAPE":              VK_ESCAPE,
	"CONVERT":             VK_CONVERT,
	"NONCONVERT":          VK_NONCONVERT,
	"ACCEPT":              VK_ACCEPT,
	"MODECHANGE":          VK_MODECHANGE,
	"SPACE":               VK_SPACE,
	"PRIOR":               VK_PRIOR,
	"NEXT":                VK_NEXT,
	"END":                 VK_END,
	"HOME":                VK_HOME,
	"LEFT":                VK_LEFT,
	"UP":                  VK_UP,
	"RIGHT":               VK_RIGHT,
	"DOWN":                VK_DOWN,
	"SELECT":              VK_SELECT,
	"PRINT":               VK_PRINT,
	"EXECUTE":             VK_EXECUTE,
	"SNAPSHOT":            VK_SNAPSHOT,
	"INSERT":              VK_INSERT,
	"DELETE":              VK_DELETE,
	"HELP":                VK_HELP,
	"LWIN":                VK_LWIN,
	"RWIN":                VK_RWIN,
	"APPS":                VK_APPS,
	"SLEEP":               VK_SLEEP,
	"NUMPAD0":             VK_NUMPAD0,
	"NUMPAD1":             VK_NUMPAD1,
	"NUMPAD2":             VK_NUMPAD2,
	"NUMPAD3":             VK_NUMPAD3,
	"NUMPAD4":             VK_NUMPAD4,
	"NUMPAD5":             VK_NUMPAD5,
	"NUMPAD6":             VK_NUMPAD6,
	"NUMPAD7":             VK_NUMPAD7,
	"NUMPAD8":             VK_NUMPAD8,
	"NUMPAD9":             VK_NUMPAD9,
	"MULTIPLY":            VK_MULTIPLY,
	"ADD":                 VK_ADD,
	"SEPARATOR":           VK_SEPARATOR,
	"SUBTRACT":            VK_SUBTRACT,
	"DECIMAL":             VK_DECIMAL,
	"DIVIDE":              VK_DIVIDE,
	"F1":                  VK_F1,
	"F2":                  VK_F2,
	"F3":                  VK_F3,
	"F4":                  VK_F4,
	"F5":                  VK_F5,
	"F6":                  VK_F6,
	"F7":                  VK_F7,
	"F8":                  VK_F8,
	"F9":                  VK_F9,
	"F10":                 VK_F10,
	"F11":                 VK_F11,
	"F12":                 VK_F12,
	"F13":                 VK_F13,
	"F14":                 VK_F14,
	"F15":                 VK_F15,
	"F16":                 VK_F16,
	"F17":                 VK_F17,
	"F18":                 VK_F18,
	"F19":                 VK_F19,
	"F20":                 VK_F20,
	"F21":                 VK_F21,
	"F22":                 VK_F22,
	"F23":                 VK_F23,
	"F24":                 VK_F24,
	"NUMLOCK":             VK_NUMLOCK,
	"SCROLL":              VK_SCROLL,
	"LSHIFT":              VK_LSHIFT,
	"RSHIFT":              VK_RSHIFT,
	"LCONTROL":            VK_LCONTROL,
	"RCONTROL":            VK_RCONTROL,
	"LMENU":               VK_LMENU,
	"RMENU":               VK_RMENU,
	"BROWSER_BACK":        VK_BROWSER_BACK,
	"BROWSER_FORWARD":     VK_BROWSER_FORWARD,
	"BROWSER_REFRESH":     VK_BROWSER_REFRESH,
	"BROWSER_STOP":        VK_BROWSER_STOP,
	"BROWSER_SEARCH":      VK_BROWSER_SEARCH,
	"BROWSER_FAVORITES":   VK_BROWSER_FAVORITES,
	"BROWSER_HOME":        VK_BROWSER_HOME,
	"VOLUME_MUTE":         VK_VOLUME_MUTE,
	"VOLUME_DOWN":         VK_VOLUME_DOWN,
	"VOLUME_UP":           VK_VOLUME_UP,
	"MEDIA_NEXT_TRACK":    VK_MEDIA_NEXT_TRACK,
	"MEDIA_PREV_TRACK":    VK_MEDIA_PREV_TRACK,
	"MEDIA_STOP":          VK_MEDIA_STOP,
	"MEDIA_PLAY_PAUSE":    VK_MEDIA_PLAY_PAUSE,
	"LAUNCH_MAIL":         VK_LAUNCH_MAIL,
	"LAUNCH_MEDIA_SELECT": VK_LAUNCH_MEDIA_SELECT,
	"LAUNCH_APP1":         VK_LAUNCH_APP1,
	"LAUNCH_APP2":         VK_LAUNCH_APP2,
	"OEM_1":               VK_OEM_1,
	"OEM_PLUS":            VK_OEM_PLUS,
	"OEM_COMMA":           VK_OEM_COMMA,
	"OEM_MINUS":           VK_OEM_MINUS,
	"OEM_PERIOD":          VK_OEM_PERIOD,
	"OEM_2":               VK_OEM_2,
	"OEM_3":               VK_OEM_3,
	"OEM_4":               VK_OEM_4,
	"OEM_5":               VK_OEM_5,
	"OEM_6":               VK_OEM_6,
	"OEM_7":               VK_OEM_7,
	"OEM_8":               VK_OEM_8,
	"OEM_102":             VK_OEM_102,
	"PROCESSKEY":          VK_PROCESSKEY,
	"PACKET":              VK_PACKET,
	"ATTN":                VK_ATTN,
	"CRSEL":               VK_CRSEL,
	"EXSEL":               VK_EXSEL,
	"EREOF":               VK_EREOF,
	"PLAY":                VK_PLAY,
	"ZOOM":                VK_ZOOM,
	"NONAME":              VK_NONAME,
	"PA1":                 VK_PA1,
	"OEM_CLEAR":           VK_OEM_CLEAR,
}

func Translate(c byte) (buf []Key) {
	buf = make([]Key, 0, 4)
	shift := false
	if Shifted(c) {
		shift = true
	}
	vk := keymap[c]
	if shift {
		buf = append(buf, Key{
			kind: KindKeyboard,
			vk:   VK_RSHIFT,
		})
	}
	buf = append(buf, []Key{
		Key{
			kind: KindKeyboard,
			vk:   vk,
		},
		Key{
			kind:  KindKeyboard,
			vk:    vk,
			flags: FlagUp,
		},
	}...)
	if shift {
		buf = append(buf, Key{
			kind:  KindKeyboard,
			vk:    VK_RSHIFT,
			flags: FlagUp,
		})
	}
	return
}

type BraceExpr struct {
	key    string
	repeat int
}

func parseBraceExpr(b []byte) (expr BraceExpr, advance int, err error) {
	if len(b) == 0 {
		return expr, 0, fmt.Errorf("bad brace expression")
	}
	if b[0] != '{' {
		return expr, 0, fmt.Errorf("missing opening brace")
	}
	advance++
	for b[advance] == ' ' && advance < len(b) {
		advance++
	}
	for j := advance; j < len(b); j++ {
		if b[j] == '}' {
			expr.key = string(b[advance:j])
			advance = j
			return
		}
	}
	return expr, advance, fmt.Errorf("missing closing brace")
}

func Shifted(c byte) bool {
	const list = "~@#$%^&*():<>?+ABCDEFGHIJKLMNOPQRSTUVWXYZ\""
	for _, d := range list {
		if byte(d) == c {
			return true
		}
	}
	return false
}

func init() {
	keymap = make(map[byte]VK)
	for c := byte(0); c < 255; c++ {
		if c >= 0x61 && c <= 0x7a {
			keymap[c] = VK(c - 0x20)
		} else {
			keymap[c] = VK(c)
		}
	}
	for k, v := range partial {
		keymap[k] = v
	}
	for k, v := range keynames {
		keynames[strings.ToLower(k)] = v
	}
}

type VK uint16

var keymap map[byte]VK

var partial = map[byte]VK{
	'@':  VK('2'),
	'#':  VK('3'),
	'$':  VK('4'),
	'%':  VK('5'),
	'^':  VK('6'),
	'&':  VK('7'),
	'*':  VK('8'),
	'(':  VK('9'),
	')':  VK('0'),
	':':  VK(';'),
	'<':  VK(','),
	'+':  VK_OEM_PLUS,
	' ':  VK_SPACE,
	'\t': VK_TAB,
	'\n': VK_RETURN,
	'=':  VK_OEM_PLUS,
	'_':  VK_TAB,
	'-':  VK_OEM_MINUS,
	'/':  VK_DIVIDE,
	'.':  VK_OEM_PERIOD,
	',':  VK_OEM_COMMA,
	'?':  VK_OEM_2,
	'>':  VK_OEM_102,
	'~':  VK_OEM_3,
	'{':  VK_OEM_4,
	'|':  VK_OEM_5,
	'}':  VK_OEM_6,
	'\'': VK_OEM_7,
	'"':  VK_OEM_7,
	'!':  VK_OEM_8,
}

const (
	VK_LBUTTON             VK = 0x01
	VK_RBUTTON             VK = 0x02
	VK_CANCEL              VK = 0x03
	VK_MBUTTON             VK = 0x04
	VK_XBUTTON1            VK = 0x05
	VK_XBUTTON2            VK = 0x06
	VK_BACK                VK = 0x08
	VK_TAB                 VK = 0x09
	VK_CLEAR               VK = 0x0C
	VK_RETURN              VK = 0x0D
	VK_SHIFT               VK = 0x10
	VK_CONTROL             VK = 0x11
	VK_MENU                VK = 0x12
	VK_PAUSE               VK = 0x13
	VK_CAPITAL             VK = 0x14
	VK_KANA                VK = 0x15
	VK_HANGUEL             VK = 0x15
	VK_HANGUL              VK = 0x16
	VK_JUNJA               VK = 0x17
	VK_FINAL               VK = 0x18
	VK_KANJI               VK = 0x19
	VK_ESCAPE              VK = 0x1B
	VK_CONVERT             VK = 0x1C
	VK_NONCONVERT          VK = 0x1D
	VK_ACCEPT              VK = 0x1E
	VK_MODECHANGE          VK = 0x1F
	VK_SPACE               VK = 0x20
	VK_PRIOR               VK = 0x21
	VK_NEXT                VK = 0x22
	VK_END                 VK = 0x23
	VK_HOME                VK = 0x24
	VK_LEFT                VK = 0x25
	VK_UP                  VK = 0x26
	VK_RIGHT               VK = 0x27
	VK_DOWN                VK = 0x28
	VK_SELECT              VK = 0x29
	VK_PRINT               VK = 0x2A
	VK_EXECUTE             VK = 0x2B
	VK_SNAPSHOT            VK = 0x2C
	VK_INSERT              VK = 0x2D
	VK_DELETE              VK = 0x2E
	VK_HELP                VK = 0x2F
	VK_LWIN                VK = 0x5B
	VK_RWIN                VK = 0x5C
	VK_APPS                VK = 0x5D
	VK_SLEEP               VK = 0x5F
	VK_NUMPAD0             VK = 0x60
	VK_NUMPAD1             VK = 0x61
	VK_NUMPAD2             VK = 0x62
	VK_NUMPAD3             VK = 0x63
	VK_NUMPAD4             VK = 0x64
	VK_NUMPAD5             VK = 0x65
	VK_NUMPAD6             VK = 0x66
	VK_NUMPAD7             VK = 0x67
	VK_NUMPAD8             VK = 0x68
	VK_NUMPAD9             VK = 0x69
	VK_MULTIPLY            VK = 0x6A
	VK_ADD                 VK = 0x6B
	VK_SEPARATOR           VK = 0x6C
	VK_SUBTRACT            VK = 0x6D
	VK_DECIMAL             VK = 0x6E
	VK_DIVIDE              VK = 0x6F
	VK_F1                  VK = 0x70
	VK_F2                  VK = 0x71
	VK_F3                  VK = 0x72
	VK_F4                  VK = 0x73
	VK_F5                  VK = 0x74
	VK_F6                  VK = 0x75
	VK_F7                  VK = 0x76
	VK_F8                  VK = 0x77
	VK_F9                  VK = 0x78
	VK_F10                 VK = 0x79
	VK_F11                 VK = 0x7A
	VK_F12                 VK = 0x7B
	VK_F13                 VK = 0x7C
	VK_F14                 VK = 0x7D
	VK_F15                 VK = 0x7E
	VK_F16                 VK = 0x7F
	VK_F17                 VK = 0x80
	VK_F18                 VK = 0x81
	VK_F19                 VK = 0x82
	VK_F20                 VK = 0x83
	VK_F21                 VK = 0x84
	VK_F22                 VK = 0x85
	VK_F23                 VK = 0x86
	VK_F24                 VK = 0x87
	VK_NUMLOCK             VK = 0x90
	VK_SCROLL              VK = 0x91
	VK_LSHIFT              VK = 0xA0
	VK_RSHIFT              VK = 0xA1
	VK_LCONTROL            VK = 0xA2
	VK_RCONTROL            VK = 0xA3
	VK_LMENU               VK = 0xA4
	VK_RMENU               VK = 0xA5
	VK_BROWSER_BACK        VK = 0xA6
	VK_BROWSER_FORWARD     VK = 0xA7
	VK_BROWSER_REFRESH     VK = 0xA8
	VK_BROWSER_STOP        VK = 0xA9
	VK_BROWSER_SEARCH      VK = 0xAA
	VK_BROWSER_FAVORITES   VK = 0xAB
	VK_BROWSER_HOME        VK = 0xAC
	VK_VOLUME_MUTE         VK = 0xAD
	VK_VOLUME_DOWN         VK = 0xAE
	VK_VOLUME_UP           VK = 0xAF
	VK_MEDIA_NEXT_TRACK    VK = 0xB0
	VK_MEDIA_PREV_TRACK    VK = 0xB1
	VK_MEDIA_STOP          VK = 0xB2
	VK_MEDIA_PLAY_PAUSE    VK = 0xB3
	VK_LAUNCH_MAIL         VK = 0xB4
	VK_LAUNCH_MEDIA_SELECT VK = 0xB5
	VK_LAUNCH_APP1         VK = 0xB6
	VK_LAUNCH_APP2         VK = 0xB7
	VK_OEM_1               VK = 0xBA
	VK_OEM_PLUS            VK = 0xBB
	VK_OEM_COMMA           VK = 0xBC
	VK_OEM_MINUS           VK = 0xBD
	VK_OEM_PERIOD          VK = 0xBE
	VK_OEM_2               VK = 0xBF
	VK_OEM_3               VK = 0xC0
	VK_OEM_4               VK = 0xDB
	VK_OEM_5               VK = 0xDC
	VK_OEM_6               VK = 0xDD
	VK_OEM_7               VK = 0xDE
	VK_OEM_8               VK = 0xDF
	VK_OEM_102             VK = 0xE2
	VK_PROCESSKEY          VK = 0xE5
	VK_PACKET              VK = 0xE7
	VK_ATTN                VK = 0xF6
	VK_CRSEL               VK = 0xF7
	VK_EXSEL               VK = 0xF8
	VK_EREOF               VK = 0xF9
	VK_PLAY                VK = 0xFA
	VK_ZOOM                VK = 0xFB
	VK_NONAME              VK = 0xFC
	VK_PA1                 VK = 0xFD
	VK_OEM_CLEAR           VK = 0xFE
)
