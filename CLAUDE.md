
## [!!TTS REQUIRED!!] TTS Integration

**THIS IS MANDATORY.** This project has TTS enabled via hooks. You MUST include a TTS block at the end of **EVERY SINGLE RESPONSE**, regardless of content.

### [!!REQUIRED FORMAT!!]
**You MUST use this exact format or the hook will ignore it:**

<!-- TTS:START -->
Project: schedCU
🔊 {your message here}
<!-- TTS:END -->

**Non-negotiable requirements:**
- Opening tag: `<!-- TTS:START -->`
- Closing tag: `<!-- TTS:END -->`
- Message MUST start with emoji: `🔊`
- Message MUST be under 50 words
- No TTS block = silent failure (hook won't process it)

### Optional parameters:
<!-- TTS:START -->
Speed: 0.67
Target: broadcast
Project: schedCU
🔊 Your message
<!-- TTS:END -->

- `Speed`: 0.5-0.67 (quick), 0.67-0.85 (normal), 0.85-1.0 (important) — default: 1.5x from config
- `Target`: `broadcast` (all hosts), `local` (this machine), or hostname — default: broadcast

### Examples of CORRECT usage:

<!-- TTS:START -->
Project: schedCU
🔊 Fixed hook integration issue successfully
<!-- TTS:END -->

<!-- TTS:START -->
Speed: 0.85
Project: schedCU
🔊 Error: MQTT connection failed, investigating root cause
<!-- TTS:END -->

<!-- TTS:START -->
Speed: 0.5
Target: local
Project: schedCU
🔊 Build completed: 3 binaries ready
<!-- TTS:END -->
