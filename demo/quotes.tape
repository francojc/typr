# quotes.tape — demonstrates typr quote mode with attribution
# Run from project root: vhs demo/quotes.tape

Output demo/quotes.gif

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 500
Set Theme "Dracula"
Set Framerate 24
Set PlaybackSpeed 1.0
Set TypingSpeed 60ms

Sleep 500ms

# Show the command — uses the bundled demo quote file
Type "./bin/typr -quotefile demo/quote-demo.json"
Sleep 400ms
Enter
Sleep 900ms

# Type the quote text:
# "Error is discipline through which we advance."
Set TypingSpeed 140ms
Type "Error "
Type "is "
Type "discipline "
Type "through "
Type "which "
Type "we "
Type "advance."

# Results screen with attribution shown — pause to display
Sleep 3s

# Exit
Escape
Sleep 400ms
