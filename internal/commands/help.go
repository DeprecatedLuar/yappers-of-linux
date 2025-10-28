package commands

import "fmt"

func Help() {
	help := `usage: yap <command> [options]

commands:
  start [options]     Start voice typing (default: --model tiny)
  toggle [options]    Smart pause/resume/start
  pause               Pause listening
  resume              Resume listening
  stop (kill)         Stop voice typing
  models              Show installed models

options:
  --model X           Model size: tiny, base, small, medium, large
  --device X          Device: cpu, cuda
  --language X        Language code (default: en)
  --tcp [PORT]        Enable TCP server (default port: 12322)
  --fast              Use fast mode (int8, less accurate but faster)

modes:
  default             Accurate mode (float32, better quality)
  --fast              Fast mode (int8, lower quality but faster)

models automatically download on first use`

	fmt.Println(help)
}
