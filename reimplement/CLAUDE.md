- never write python.  try hard to use go wherever possible, including to replace bash scripts

## [!!TTS REQUIRED!!] TTS Integration

**THIS IS MANDATORY.** This project has TTS enabled via hooks. You MUST include a TTS block at the end of **EVERY SINGLE RESPONSE**, regardless of content.

### [!!REQUIRED FORMAT!!]
**You MUST use this exact format or the hook will ignore it:**

<!-- TTS:START -->
Project: reimplement
🔊 {your message here}
<!-- TTS:END -->

**Non-negotiable requirements:**
- Opening tag: `<!-- TTS:START -->`
- Closing tag: `<!-- TTS:END -->`
- Message MUST start with emoji: `🔊`
- Message should be 1-7 words (absolutely no more than 10 words)
- No TTS block = silent failure (hook will not process it)

### Message Verbosity Guidelines

**BE TERSE.** These are audio notifications, not summaries.

**Success/Completion** (2-4 words):
- ✓ "Tests passing"
- ✓ "Build complete"
- ✓ "Deployed"
- ✗ "All tests have passed and the application is ready" (too verbose)

**Errors/Warnings** (3-5 words):
- ✓ "Build failed"
- ✓ "Connection error"
- ✓ "Test timeout"
- ✗ "Error: something failed, investigating root cause" (too verbose)

**Informational** (3-5 words):
- ✓ "Ready to deploy"
- ✓ "Fix verified"
- ✓ "Changes committed"
- ✗ "Found and completed the fix process successfully" (too verbose)

### Optional Parameters

<!-- TTS:START -->
Target: broadcast
Project: reimplement
🔊 Your message
<!-- TTS:END -->

- **DO NOT change Speed** - Always use default speed from config
- `Target`: `broadcast` (all hosts), `local` (this machine), or hostname — default: broadcast
  - `broadcast` = message plays on ALL machines with tts-sink running
  - `local` = message plays ONLY on this machine
  - hostname = message plays on that specific machine

### Examples of CORRECT Usage

<!-- TTS:START -->
Project: reimplement
🔊 Build complete
<!-- TTS:END -->

<!-- TTS:START -->
Project: reimplement
🔊 Error detected
<!-- TTS:END -->

<!-- TTS:START -->
Target: local
Project: reimplement
🔊 Done
<!-- TTS:END -->
