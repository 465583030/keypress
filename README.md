# keypress
Command line keystroke pipeline for Windows

- keypress reads `text` from standard input

    echo `hello world `| keypress

- `text` represents virtual keys

## Three types

### Literal text
    echo hello world | keypress

### Bracket Expressions
    echo {TAB}{DOWN}{UP}| keypress

### Chords
    echo #r|keypress
   
  
  
## Reference: Chords and Buttons
    ^   Holds CTRL for next key
    #   Holds Windows button for next key
    !   Holds ALT for next key
    +   Hold shift for next key
    
   ```{LBUTTON}
   {LBUTTON}
   {RBUTTON}
   {CANCEL}
   {MBUTTON}
   {XBUTTON1}
   {XBUTTON2}
   {BACK}
   {TAB}
   {CLEAR}
   {RETURN}
   {SHIFT}
   {CONTROL}
   {MENU}
   {PAUSE}
   {CAPITAL}
   {KANA}
   {HANGUEL}
   {HANGUL}
   {JUNJA}
   {FINAL}
   {KANJI}
   {ESCAPE}
   {CONVERT}
   {NONCONVERT}
   {ACCEPT}
   {MODECHANGE}
   {SPACE}
   {PRIOR}
   {NEXT}
   {END}
   {HOME}
   {LEFT}
   {UP}
   {RIGHT}
   {DOWN}
   {SELECT}
   {PRINT}
   {EXECUTE}
   {SNAPSHOT}
   {INSERT}
   {DELETE}
   {HELP}
   {LWIN}
   {RWIN}
   {APPS}
   {SLEEP}
   {NUMPAD0}
   {NUMPAD1}
   {NUMPAD2}
   {NUMPAD3}
   {NUMPAD4}
   {NUMPAD5}
   {NUMPAD6}
   {NUMPAD7}
   {NUMPAD8}
   {NUMPAD9}
   {MULTIPLY}
   {ADD}
   {SEPARATOR}
   {SUBTRACT}
   {DECIMAL}
   {DIVIDE}
   {F1}
   {F2}
   {F3}
   {F4}
   {F5}
   {F6}
   {F7}
   {F8}
   {F9}
   {F10}
   {F11}
   {F12}
   {F13}
   {F14}
   {F15}
   {F16}
   {F17}
   {F18}
   {F19}
   {F20}
   {F21}
   {F22}
   {F23}
   {F24}
   {NUMLOCK}
   {SCROLL}
   {LSHIFT}
   {RSHIFT}
   {LCONTROL}
   {RCONTROL}
   {LMENU}
   {RMENU}
   {BROWSER_BACK}
   {BROWSER_FORWARD}
   {BROWSER_REFRESH}
   {BROWSER_STOP}
   {BROWSER_SEARCH}
   {BROWSER_FAVORITES}
   {BROWSER_HOME}
   {VOLUME_MUTE}
   {VOLUME_DOWN}
   {VOLUME_UP}
   {MEDIA_NEXT_TRACK}
   {MEDIA_PREV_TRACK}
   {MEDIA_STOP}
   {MEDIA_PLAY_PAUSE}
   {LAUNCH_MAIL}
   {LAUNCH_MEDIA_SELECT}
   {LAUNCH_APP1}
   {LAUNCH_APP2}
   {OEM_1}
   {OEM_PLUS}
   {OEM_COMMA}
   {OEM_MINUS}
   {OEM_PERIOD}
   {OEM_2}
   {OEM_3}
   {OEM_4}
   {OEM_5}
   {OEM_6}
   {OEM_7}
   {OEM_8}
   {OEM_102}
   {PROCESSKEY}
   {PACKET}
   {ATTN}
   {CRSEL}
   {EXSEL}
   {EREOF}
   {PLAY}
   {ZOOM}
