# Use Volcengine V3 for Lookup Text Pronunciation

Moon CLI will use the Volcengine/Doubao Speech V3 TTS API family for lookup text pronunciation, but it does not require the bidirectional WebSocket API for `mo vb` because the lookup text is already complete before synthesis starts. This keeps the door open for current V3 voices and lower-latency APIs while avoiding bidirectional streaming complexity unless a future command needs realtime text input.
