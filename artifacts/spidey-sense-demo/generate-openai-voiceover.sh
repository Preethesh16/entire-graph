#!/usr/bin/env bash
set -euo pipefail

if [[ -z ${OPENAI_API_KEY:-} ]]; then
  echo "OPENAI_API_KEY is required" >&2
  exit 1
fi

script_dir=$(cd "$(dirname "$0")" && pwd)
jq -n \
  --rawfile narration "$script_dir/narration.txt" \
  '{model:"gpt-4o-mini-tts",voice:"marin",instructions:"Confident, warm product narrator. Natural pacing, clear technical terms, subtle excitement, no exaggerated sales tone.",input:$narration,response_format:"wav"}' \
  | curl --fail --silent --show-error https://api.openai.com/v1/audio/speech \
      -H "Authorization: Bearer $OPENAI_API_KEY" \
      -H 'Content-Type: application/json' \
      --data-binary @- \
      --output "$script_dir/openai-voiceover.wav"
