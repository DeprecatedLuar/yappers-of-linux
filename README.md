<H1 align="center">
  Yappers of Linux
</H1>

<p align="center">
  <img src="other/assets/yappers-of-linux-banner.png"/>
</p>

<p align="center">Voice typing for Linux that doesn't suck</p>

<p align="center">
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/stargazers">
    <img src="https://img.shields.io/github/stars/DeprecatedLuar/yappers-of-linux?style=for-the-badge&logo=github&color=1f6feb&logoColor=white&labelColor=black"/>
  </a>
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/releases">
    <img src="https://img.shields.io/github/v/release/DeprecatedLuar/yappers-of-linux?style=for-the-badge&logo=go&color=00ADD8&logoColor=white&labelColor=black"/>
  </a>
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/DeprecatedLuar/yappers-of-linux?style=for-the-badge&color=green&labelColor=black"/>
  </a>
</p>

---

## Somehow I had to build this

(Local) Voice typing on Linux either doesn't work or was made in the past century. How android beats Linux on that for the past 10 years? I can't let that slide.

---

## The cool features you've never seen before

- **Single binary** - The backend is Python but I managed to embed everything in a single binary, which means runs anywhere with no setup (well, at least Linux)
- **Performance modes** - Faster or more accurate modes based on your hardware
- **One toggle** - `yap toggle` pauses, resumes, or starts (control via cli)
- **TCP server** - It can serve its state over TCP for status bars, widgets, AI Girlfriends, etc.
- **Output file** - Pipe transcriptions to other scripts for automation or custom scripts
- **Actually private** - Literally Whisper (Wow!)
- **Configurable** - Change models, languages, devices, and all that stuff

---

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/DeprecatedLuar/yappers-of-linux/main/install.sh | bash
```

<details>
<summary>Other Install Methods</summary>

<br>

**Manual Install**
```bash
# Download binary from releases (amd64)
wget https://github.com/DeprecatedLuar/yappers-of-linux/releases/latest/download/yap-linux-amd64
chmod +x yap-linux-amd64
sudo mv yap-linux-amd64 /usr/local/bin/yap
yap start  # Auto-installs everything

# Or for arm64
wget https://github.com/DeprecatedLuar/yappers-of-linux/releases/latest/download/yap-linux-arm64
chmod +x yap-linux-arm64
sudo mv yap-linux-arm64 /usr/local/bin/yap
```

**Build From Source**
```bash
git clone https://github.com/DeprecatedLuar/yappers-of-linux.git
cd yappers-of-linux
go build -o yap cmd/main.go
./yap start
```

**System Requirements**:
- `python3` (3.10+)
- `portaudio19-dev` (for mic access)
- `ydotool` + `ydotoold` (Wayland) OR `xdotool` (X11)

</details>

First run takes ~2 minutes to download and set up everything. After that, it's instant.

---

## Commands

| Command | Arguments         | Description                                      |
|---------|-------------------|--------------------------------------------------|
| start   | `[options]`       | Start voice typing                               |
| stop    |                   | Stop voice typing                                |
| toggle  |                   | Smart pause/resume/start                         |
| pause   |                   | Pause listening                                  |
| resume  |                   | Resume listening                                 |
| output  |                   | View output file (aliases: log, cat, show)       |
| models  |                   | Show installed models                            |
| config  |                   | Open config in editor                            |
| help    | `[topic]`         | Show help information                            |

<details>
<summary>Flags</summary>

<br>

| Flag                  | Description                                      |
|-----------------------|--------------------------------------------------|
| `--model MODEL`       | Choose model (tiny/base/small/medium/large)      |
| `--device DEVICE`     | Use cpu or cuda                                  |
| `--language LANG`     | Set language (en/es/fr/etc)                      |
| `--tcp [PORT]`        | Enable TCP server (default: 12322)               |
| `--fast`              | Fast mode (int8, less accurate)                  |
| `--no-typing`         | Print to terminal only, don't type               |

</details>

## Usage

```bash
yap start                     # Start listening
yap start --model small       # Use better model
yap start --fast              # Faster but less accurate
yap toggle                    # Pause/resume/start
yap stop                      # Stop
```

<details>
<summary>Available Models</summary>

<br>

Models auto-download on first use:

| Model  | Size   | Speed      | Accuracy |
|--------|--------|------------|----------|
| tiny   | ~75MB  | Fastest    | Basic    |
| base   | ~150MB | Fast       | Good     |
| small  | ~500MB | Balanced   | Better   |
| medium | ~1.5GB | Slow       | Great    |
| large  | ~3GB   | Slowest    | Best     |

</details>

<details>
<summary>More stuff you can do</summary>

<br>

```bash
yap start --device cuda       # Use GPU instead
yap start --language es       # Spanish (or any other language)
yap start --tcp               # Enable state server on port 12322
yap start --no-typing         # Just prints to terminal, doesn't type
yap models                    # See what models you have
yap config                    # Open config in your editor
```

</details>

<details>
<summary>Configuration</summary>

<br>

Config file lives at `~/.config/yappers-of-linux/config.toml` and gets created on first run.

```toml
notifications = "start,urgent"   # When to notify you
model = "tiny"                   # Which model to use
device = "cpu"                   # cpu or cuda
language = "en"                  # What language you're speaking
fast_mode = false                # Trade accuracy for speed
enable_typing = true             # Type into active window
output_file = false              # Write to output.txt for piping/automation
```

Run `yap help config` if you want all the details.

</details>


<details>
<summary>State Monitoring (if you're into that)</summary>

<br>

Want to plug this into your status bar or a desktop widget?

```bash
yap start --tcp        # Starts on port 12322
nc 127.0.0.1 12322     # Test it out
```

Spits out JSON with the current state. Inspired by [Kanata's TCP port](https://github.com/jtroo/kanata).

</details>

<details>
<summary>Output File (for automation)</summary>

<br>

Enable `output_file = true` in config to write transcriptions to `~/.config/yappers-of-linux/output.txt`.

**How it works**:
- File is ephemeral - deleted on each `yap start` (fresh session)
- Each transcription is separated by a blank line (paragraph style)
- View anytime with `yap output` (or `yap log`, `yap cat`, `yap show`)

**Use cases**:
```bash
# View the output file
yap output

# Pipe to another script
tail -f ~/.config/yappers-of-linux/output.txt | your-script.sh

# Process with jq/awk/whatever
cat ~/.config/yappers-of-linux/output.txt | process-commands

# Voice-controlled automation
while read line; do handle_command "$line"; done < output.txt
```

</details>

---

## How it works

<img src="other/assets/ermactually.jpeg" alt="Actually..." align="right" width="200"/>

Constantly records audio to a 1.5-second circular buffer. When VAD detects speech, it captures that buffer plus whatever you continue saying. After 0.8 seconds of silence, it sends everything to Whisper for transcription and types the result. This means your opening words are never lost.

Filters out hallucinations using confidence thresholds - keyboard clicks and background noise won't turn into random words.

The entire engine (Python + dependencies) is embedded in a single Go binary that self-extracts and manages its own environment. First run takes ~2 minutes to set up, after that it's instant.

---

<p align="center">
  <a href="https://github.com/DeprecatedLuar/yappers-of-linux/issues">
    <img src="https://img.shields.io/badge/Found%20a%20bug%3F-Report%20it!-red?style=for-the-badge&logo=github&logoColor=white&labelColor=black"/>
  </a>
</p>
