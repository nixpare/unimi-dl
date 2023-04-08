import os
import sys
import urllib.request as req

video_manifest = sys.argv[1]
video_name = sys.argv[2]

manifest_content = str(req.urlopen(video_manifest).read(), encoding='utf-8')

video_base_url = video_manifest[:video_manifest.rindex("/")]
video_path = os.getcwd() + "/" + video_name + ".ts"

segment_manifest_url = video_base_url + "/" + manifest_content.rstrip(" \n").split("\n")[-1]
segment_manifest_content = str(req.urlopen(segment_manifest_url).read(), encoding='utf-8')

segment_urls = segment_manifest_content.split("\n")

with open(video_path, 'wb') as video_file:
    for line in segment_urls:
        if (line.startswith("#") or line == ""):
            continue
        
        video_seg = req.urlopen(video_base_url + "/" + line).read()
        video_file.write(video_seg)

print("Exited")
