# DVD Screensaver

A Go implementation of the classic DVD screensaver animation where the DVD logo bounces around the screen, changing color when it hits the edges.

## Features

- Realistic DVD logo with rounded corners and shadow effect
- Random starting position and movement direction
- Color changes on edge collision
- Smooth animation using the Ebiten game library

## Requirements

- Go 1.20 or higher
- Ebiten v2 library

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/yourusername/dvd-screensaver.git
   cd dvd-screensaver
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Running the Screensaver

Simply run:

```
go run .
```

## How It Works

The program creates a window with a black background and displays a DVD logo that moves around the screen. When the logo hits any edge of the screen, it bounces off at the same angle and changes to a new color from a predefined set of colors.

The animation uses the Ebiten game library, which provides a simple and efficient way to create 2D games and animations in Go.

## Customization

You can customize the screensaver by modifying the following constants in `main.go`:

- `screenWidth` and `screenHeight`: Change the size of the window
- `logoWidth` and `logoHeight`: Change the size of the DVD logo
- `logoColors`: Add or remove colors from the color cycle

## License

MIT