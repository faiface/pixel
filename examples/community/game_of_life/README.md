# Conway's Game of Lfe

Created by [Nathan Leniz](https://github.com/terakilobyte).
Inspired by and heavily uses [the doc](https://golang.org/doc/play/life.go)

> The Game of Life, also known simply as Life, is a cellular automaton devised by the British mathematician John Horton Conway in 1970. The "game" is a zero-player game, meaning that its evolution is determined by its initial state, requiring no further input. One interacts with the Game of Life by creating an initial configuration and observing how it evolves, or, for advanced "players", by creating patterns with particular properties. The Game has been reprogrammed multiple times in various coding languages.

For more information, please see the [wikipedia](https://en.wikipedia.org/wiki/Conway%27s_Game_of_Life) article.

## Use

    go run main.go -h
      -frameRate duration
          The framerate in milliseconds (default 33ms)
      -size int
          The size of each cell (default 5)
      -windowSize float
          The pixel size of one side of the grid (default 800)

![life](life.png)
