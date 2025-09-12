package progress

import (
	"context"
	"time"
	"fmt"

	"github.com/fatih/color"
)

func ShowSpinnerProgress(ctx context.Context) {
	indicators := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    i := 0

	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// animation stop and clear the line
			// fmt.Print("\r\033[K")
			return
		case <-ticker.C:
			fmt.Printf("\r🤖 AI: %s Thinking...", indicators[i%len(indicators)])
			i++
		}
	}
}

func ShowColorfulProgress(ctx context.Context) {
    colors := []*color.Color{
        color.New(color.FgRed),
        color.New(color.FgYellow), 
        color.New(color.FgGreen),
        color.New(color.FgCyan),
        color.New(color.FgBlue),
        color.New(color.FgMagenta),
    }
	indicators := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

    i := 0
    ticker := time.NewTicker(120 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            // fmt.Print("\r")
            return
        case <-ticker.C:
            fmt.Print("\r🤖 ")
            // colors[i%len(colors)].Printf("AI: ✨ Generating response...")
            colors[i%len(colors)].Printf("\r🤖 AI: %s Thinking...", indicators[i%len(indicators)])
            i++
        }
    }
}
