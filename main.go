package main

import (
	"flag"
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"time"

	"github.com/massung/chip-8/chip8"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	/// The file to load as a ROM or assemble.
	///
	File string

	/// True if the File is assembly code.
	///
	Assemble bool

	/// True if the ROM should load paused.
	///
	Break bool

	/// The CHIP-8 virtual machine.
	///
	VM *chip8.CHIP_8

	/// The SDL Window and Renderer.
	///
	Window *sdl.Window
	Renderer *sdl.Renderer
)

func init() {
	runtime.LockOSThread()
}

func main() {
	var err error

	// seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// parse the command line
	flag.BoolVar(&Assemble, "a", false, "Assemble file before loading.")
	flag.BoolVar(&Break, "b", false, "Start ROM paused.")
	flag.Parse()

	// get the file name of the ROM to load
	File = flag.Arg(0)

	// initialize SDL or panic
	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		panic(err)
	}

	// create the main window and renderer or panic
	flags := sdl.WINDOW_OPENGL | sdl.WINDOWPOS_CENTERED
	if Window, Renderer, err = sdl.CreateWindowAndRenderer(614, 380, uint32(flags)); err != nil {
		panic(err)
	}

	// set the icon
	if icon, err := sdl.LoadBMP("data/chip_8.bmp"); err == nil {
		mask := sdl.MapRGB(icon.Format, 255, 0, 255)

		// create the mask color key and set the icon
		icon.SetColorKey(1, mask)
		Window.SetIcon(icon)
	}

	// set the title
	Window.SetTitle("CHIP-8")

	// initialize subsystems
	InitDebug()

	// show copyright information
	fmt.Println("CHIP-8, Copyright 2017 by Jeffrey Massung")
	fmt.Println("All rights reserved")
	fmt.Println()

	// load either the file specified or default to PONG
	if File != "" {
		fmt.Println("Loading", filepath.Base(File))
	} else {
		fmt.Println("Loading PONG")
	}

	// create a new CHIP-8 virtual machine, load the ROM..
	Load()

	InitScreen()
	InitAudio()
	InitFont()

	// initially break into debugger?
	Paused = Break

	// set processor speed and refresh rate
	clock := time.NewTicker(time.Millisecond * 2)
	video := time.NewTicker(time.Second / 60)

	// notify that the main loop has started
	fmt.Println("\nStarting program; press F1 for help")

	// loop until window closed or user quit
	for ProcessEvents() {
		select {
		case <-video.C:
			Refresh()
		case <-clock.C:
			res := VM.Process(Paused)

			switch res.(type) {
			case chip8.Breakpoint:
				fmt.Println(res.Error())

				// break the emulation
				Paused = true
			}
		}
	}
}

func Load() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)

			// load the empty, dummy rom
			VM = chip8.LoadROM(chip8.Dummy)
		}
	}()

	if File == "" {
		VM = chip8.LoadROM(chip8.Pong)
	} else {
		if Assemble {
			asm := chip8.Assemble(File)

			// load the assembled program
			VM = chip8.LoadROM(asm.ROM)

			// add all the breakpoints from the assembly
			for _, b := range asm.Breakpoints {
				VM.AddBreakpoint(b)
			}
		} else {
			VM = chip8.LoadFile(File)
		}
	}
}

func Refresh() {
	Renderer.SetDrawColor(32, 42, 53, 255)
	Renderer.Clear()

	// frame various portions of the app
	Frame(8, 8, 386, 194)
	Frame(8, 208, 386, 164)
	Frame(402, 8, 204, 194)
	Frame(402, 208, 204, 164)

	// update the video screen and copy it
	RefreshScreen()
	CopyScreen(10, 10, 384, 192)

	// debug assembly and virtual registers
	DebugLog(12, 212)
	DebugAssembly(406, 11)
	DebugRegisters(406, 212)

	// show the new frame
	Renderer.Present()
}

func Frame(x, y, w, h int) {
	Renderer.SetDrawColor(0, 0, 0, 255)
	Renderer.DrawLine(x, y, x + w, y)
	Renderer.DrawLine(x, y, x, y + h)

	// highlight
	Renderer.SetDrawColor(95, 112, 120, 255)
	Renderer.DrawLine(x + w, y, x + w, y + h)
	Renderer.DrawLine(x, y + h, x + w, y + h)
}
