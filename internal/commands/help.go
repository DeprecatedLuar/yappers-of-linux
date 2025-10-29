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
  config              Open config file in $EDITOR

options:
  --model X           Model size: tiny, base, small, medium, large
  --device X          Device: cpu, gpu (or cuda)
  --language X        Language code (default: en)
  --lang X            Short alias for --language
  --tcp [PORT]        Enable TCP server (default port: 12322)
  --fast              Use fast mode (int8, less accurate but faster)
  --no-typing         Disable keyboard typing (only print to terminal)

shortcuts:
  --cpu               Alias for --device cpu
  --gpu               Alias for --device gpu
  --cuda              Alias for --device gpu

modes:
  default             Accurate mode (float32, better quality)
  --fast              Fast mode (int8, lower quality but faster)

models automatically download on first use`

	fmt.Println(help)
}
