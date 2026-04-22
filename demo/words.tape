# words.tape — demonstrates typr word/text mode
# Run from project root: vhs demo/words.tape

Output demo/words.gif

Set Shell "bash"
Set FontSize 16
Set Width 100
Set Height 28
Set Theme "Dracula"
Set Framerate 24
Set PlaybackSpeed 1.0
Set TypingSpeed 60ms

Sleep 500ms

# Show the command being typed at the shell
Type "./bin/typr demo/words-demo.txt"
Sleep 400ms
Enter
Sleep 900ms

# Type the known content from words-demo.txt:
# "the quick brown fox jumps over the lazy dog and the cat sat flat"
Set TypingSpeed 130ms
Type "the "
Type "quick "
Type "brown "
Type "fox "
Type "jumps "
Type "over "
Type "the "
Type "lazy "
Type "dog "
Type "and "
Type "the "
Type "cat "
Type "sat "
Type "flat"

# Results screen appears — pause to show it
Sleep 3s

# Exit
Escape
Sleep 400ms
