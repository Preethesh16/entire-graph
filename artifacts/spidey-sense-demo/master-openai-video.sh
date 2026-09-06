#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "$0")" && pwd)
"$script_dir/generate-openai-voiceover.sh"

ffmpeg -y \
  -i "$script_dir/raw/01-create-room.webm" \
  -i "$script_dir/raw/02-join-browser.webm" \
  -i "$script_dir/raw/03-plan-graph-activity.webm" \
  -i "$script_dir/openai-voiceover.wav" \
  -filter_complex "[0:v]fps=30,format=yuv420p[v0];[1:v]fps=30,format=yuv420p[v1];[2:v]fps=30,format=yuv420p[v2];[v0][v1][v2]concat=n=3:v=1:a=0,tpad=stop_mode=clone:stop_duration=20,subtitles='$script_dir/captions.srt':force_style='FontName=Noto Sans,FontSize=18,PrimaryColour=&H00FFFFFF,OutlineColour=&HCC000000,BorderStyle=3,Outline=1,Shadow=0,MarginV=34,Alignment=2'[video];[3:a]loudnorm=I=-16:TP=-1.5:LRA=8,afade=t=in:st=0:d=0.25[audio]" \
  -map '[video]' -map '[audio]' \
  -c:v libx264 -preset medium -crf 20 -profile:v high -level 4.1 \
  -c:a aac -b:a 192k -movflags +faststart -shortest \
  "$script_dir/spidey-sense-workflow-openai.mp4"
