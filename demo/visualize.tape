# visualize.tape — demonstrates typr visualize subcommand
# Run from project root: vhs demo/visualize.tape
# Requires mock data: run scripts/gen-mock-data.sh first

Output demo/visualize.gif

Set Shell "bash"
Set FontSize 16
Set Width 100
Set Height 34
Set Theme "Dracula"
Set Framerate 24
Set PlaybackSpeed 1.0
Set TypingSpeed 60ms

Sleep 500ms

# Generate mock data (idempotent — safe to re-run)
Exec scripts/gen-mock-data.sh

Sleep 300ms

# Show the visualize command
Type "./bin/typr visualize words-stats.csv"
Sleep 400ms
Enter
Sleep 800ms

# Graph renders immediately — pause to let the viewer read it
Sleep 5s

# Exit (visualize waits on fmt.Scanln — Enter exits)
Enter
Sleep 400ms
