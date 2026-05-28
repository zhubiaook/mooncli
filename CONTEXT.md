# Moon CLI

Moon CLI is a terminal language utility for Chinese-English translation, vocabulary learning, and English sentence checking.

## Language

**Lookup text**:
The exact word, phrase, or sentence supplied by the user to a command.
_Avoid_: Query, prompt

**Vocabulary explanation**:
The printed teaching response produced for lookup text passed to `mo vb`.
_Avoid_: Answer, result

**Lookup text pronunciation**:
Audio playback of the lookup text itself before the vocabulary explanation is printed.
_Avoid_: Text-to-speech result, spoken explanation

**Pronunciation replay**:
Multiple plays of the same synthesized lookup text pronunciation for a single lookup.
_Avoid_: Repeat synthesis, repeated explanation

**Moon CLI settings**:
The local user configuration file that stores provider credentials and defaults for Moon CLI.
_Avoid_: Claude settings, separate TTS config

**Pronunciation credential**:
The Volcengine V3 API key used to authorize lookup text pronunciation.
_Avoid_: App ID, access token

**Pronunciation voice**:
The configured Volcengine speaker used for lookup text pronunciation.
_Avoid_: Default voice, hard-coded speaker

## Example Dialogue

Developer: For `mo vb resilient`, should lookup text pronunciation speak the generated vocabulary explanation?
Domain expert: No. It should speak only the lookup text, "resilient", then print the vocabulary explanation.

Developer: For `mo vb hello --repeat 3`, should the CLI synthesize the lookup text three times?
Domain expert: No. It should synthesize once and use pronunciation replay to play the same audio three times.

Developer: Where should the pronunciation provider credentials live?
Domain expert: In Moon CLI settings, alongside the existing language model credentials.

Developer: Should lookup text pronunciation use App ID and access token?
Domain expert: No. Use a Volcengine V3 API key as the pronunciation credential.

Developer: Can the CLI pick a built-in pronunciation voice?
Domain expert: No. The pronunciation voice must be configured in Moon CLI settings.
