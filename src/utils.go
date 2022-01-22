package main

type Colour string

const (
	ColourGreen  Colour = "\033[32m"
	ColourRed    Colour = "\033[31m"
	ColourReset  Colour = "\033[0m"
	ColourYellow Colour = "\033[33m"
)
