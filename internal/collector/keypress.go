package collector

import (
	"fmt"
	"log"
	"sync"
	"time"
	"unsafe"

	"github.com/nilszeilon/devstats/internal/domain"
	"github.com/nilszeilon/devstats/internal/storage"
)

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
// #import <ApplicationServices/ApplicationServices.h>
// void external_go_callback(void*, int64_t);
//
// static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
//     if (type == kCGEventKeyDown) {
//         int64_t keycode = CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
//         external_go_callback(refcon, keycode);
//     }
//     return event;
// }
//
// static void startEventTap(void* callback) {
//     CGEventMask mask = CGEventMaskBit(kCGEventKeyDown);
//     CFMachPortRef tap = CGEventTapCreate(
//         kCGSessionEventTap,
//         kCGHeadInsertEventTap,
//         kCGEventTapOptionDefault,
//         mask,
//         eventCallback,
//         callback
//     );
//
//     if (!tap) {
//         return;
//     }
//
//     CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, tap, 0);
//     CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
//     CGEventTapEnable(tap, true);
//     CFRunLoopRun();
// }
import "C"

var (
	globalCallback *KeypressCollector
	callbackMutex  sync.Mutex
)

// KeypressCollector handles collection of keypress data
type KeypressCollector struct {
	store    storage.Store[domain.KeypressData]
	stopChan chan struct{}
	keyChan  chan int64
}

// NewKeypressCollector creates a new keypress collector
func NewKeypressCollector(store storage.Store[domain.KeypressData]) *KeypressCollector {
	return &KeypressCollector{
		store:    store,
		stopChan: make(chan struct{}),
	}
}

//export external_go_callback
func external_go_callback(_ unsafe.Pointer, keycode int64) {
	callbackMutex.Lock()
	if globalCallback != nil && globalCallback.keyChan != nil {
		globalCallback.keyChan <- keycode
	}
	callbackMutex.Unlock()
}

// keyCodeToString converts a macOS keycode to a string representation
func keyCodeToString(keycode int64) string {
	keycodeMap := map[int64]string{
		0:   "a",
		1:   "s",
		2:   "d",
		3:   "f",
		4:   "h",
		5:   "g",
		6:   "z",
		7:   "x",
		8:   "c",
		9:   "v",
		10:  "§", // Section symbol on some keyboards
		11:  "b",
		12:  "q",
		13:  "w",
		14:  "e",
		15:  "r",
		16:  "y",
		17:  "t",
		18:  "1",
		19:  "2",
		20:  "3",
		21:  "4",
		22:  "6",
		23:  "5",
		24:  "´",
		25:  "9",
		26:  "7",
		27:  "+",
		28:  "8",
		29:  "0",
		30:  "¨",
		31:  "o",
		32:  "u",
		33:  "å",
		34:  "i",
		35:  "p",
		36:  "return",
		37:  "l",
		38:  "j",
		39:  "ä",
		40:  "k",
		41:  "ö",
		42:  "'",
		43:  ",",
		44:  "-",
		45:  "n",
		46:  "m",
		47:  ".",
		48:  "tab",
		49:  "space",
		50:  "<",
		51:  "delete",
		53:  "escape",
		55:  "command",
		56:  "shift",
		57:  "capslock",
		58:  "option",
		59:  "control",
		60:  "right_shift",
		61:  "right_option",
		62:  "right_control",
		63:  "fn",
		64:  "f17",
		65:  "keypad_decimal",
		67:  "keypad_multiply",
		69:  "keypad_plus",
		71:  "keypad_clear",
		75:  "keypad_divide",
		76:  "keypad_enter",
		78:  "keypad_minus",
		79:  "f18",
		80:  "f19",
		81:  "keypad_equals",
		82:  "keypad_0",
		83:  "keypad_1",
		84:  "keypad_2",
		85:  "keypad_3",
		86:  "keypad_4",
		87:  "keypad_5",
		88:  "keypad_6",
		89:  "keypad_7",
		91:  "keypad_8",
		92:  "keypad_9",
		96:  "f5",
		97:  "f6",
		98:  "f7",
		99:  "f3",
		100: "f8",
		101: "f9",
		102: "f11",
		103: "f13",
		104: "f16",
		105: "f14",
		106: "f10",
		107: "f12",
		108: "f15",
		109: "f4",
		110: "f2",
		111: "f1",
		114: "help",
		115: "home",
		116: "page_up",
		117: "forward_delete",
		118: "f4",
		119: "end",
		120: "f2",
		121: "page_down",
		122: "f1",
		123: "left_arrow",
		124: "right_arrow",
		125: "down_arrow",
		126: "up_arrow",
	}

	if str, ok := keycodeMap[keycode]; ok {
		return str
	}
	return fmt.Sprintf("key_%d", keycode)
}

// Start begins collecting keypress data
func (kc *KeypressCollector) Start() error {
	kc.keyChan = make(chan int64, 100)

	go func() {
		for {
			select {
			case <-kc.stopChan:
				return
			case keycode := <-kc.keyChan:
				data := domain.KeypressData{
					Key:       keyCodeToString(keycode),
					Timestamp: time.Now(),
				}

				if err := kc.store.Save(data); err != nil {
					log.Printf("Error saving keypress: %v", err)
				}
			}
		}
	}()

	// Register this collector as the global callback handler
	callbackMutex.Lock()
	globalCallback = kc
	callbackMutex.Unlock()

	// Start the event tap in a separate goroutine
	go C.startEventTap(nil)

	return nil
}

// Stop stops collecting keypress data
func (kc *KeypressCollector) Stop() {
	callbackMutex.Lock()
	if globalCallback == kc {
		globalCallback = nil
	}
	callbackMutex.Unlock()
	close(kc.stopChan)
}

// Record saves a keypress event (mainly for testing)
func (kc *KeypressCollector) Record(key string) error {
	data := domain.KeypressData{
		Key:       key,
		Timestamp: time.Now(),
	}
	return kc.store.Save(data)
}
