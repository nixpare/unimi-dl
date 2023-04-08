import io
import os
import queue
import threading
import urllib.request as req

class VideoSegment:
    data: bytes
    ready: bool

    def __init__(self):
        self.data = None
        self.ready = False

class Video:
    name: str
    url: str
    segments: dict[int, VideoSegment]
    file: io.BufferedWriter
    download_queue: queue.Queue

    __concurrent_downloads__ = 8
    __flush_after__ = 4

    def __init__(self, name: str, url: str):
        self.name = name
        self.url = url
        self.segments = dict()
        self.download_queue = queue.Queue()

        video_path = os.getcwd() + "/" + self.name + ".ts"
        self.file = open(video_path, 'wb')

    def download_segment(self, index: int, url: str):
        response = req.urlopen(url)

        self.segments[index].data = response.read()
        self.segments[index].ready = True

        self.download_queue.put(item=index, block=False)

    def flush(self):
        slice_start = -1

        for i in range(len(self.segments)):
            if (self.segments[i].ready and self.segments[i].data != None):
                slice_start = i
                break

        if (slice_start == -1):
            return
        
        slice_end = slice_start
        for i in range(slice_start+1, len(self.segments)):
            if (self.segments[i].ready == False):
                break
            slice_end = i

        flush_after = max(self.__flush_after__, self.__concurrent_downloads__)
        if (slice_end - slice_start < flush_after - 1):
            return

        print(f'Flushing segments from {slice_start} to {slice_end}')
        
        for i in range(slice_start, slice_end+1):
            self.file.write(self.segments[i].data)
            self.segments[i].data = None

    def download_video(self):
        print(f"Downloading video {self.name} with url {self.url}")
        
        manifest_content = str(req.urlopen(self.url).read(), encoding='utf-8')
        video_base_url = self.url[:self.url.rindex("/")]

        segment_manifest_url = video_base_url + "/" + manifest_content.rstrip(" \n").split("\n")[-1]
        segment_manifest_content = str(req.urlopen(segment_manifest_url).read(), encoding='utf-8')

        segment_urls = segment_manifest_content.split("\n")
        for i in range(len(segment_urls)):
            self.segments[i] = VideoSegment()

        seg_threads: list[threading.Thread] = list()
        seg_counter = 0

        for line in segment_urls:
            if (line.startswith("#") or line == ""):
                continue

            seg_threads.append(threading.Thread(target=self.download_segment, args=(seg_counter, video_base_url + "/" + line)))
            seg_counter += 1

        next_downloading = 0

        while (next_downloading < seg_counter):
            if (next_downloading < self.__concurrent_downloads__):
                seg_threads[next_downloading].start()
                next_downloading += 1
                continue

            self.download_queue.get()

            seg_threads[next_downloading].start()
            next_downloading += 1

            self.flush()

        for thread in seg_threads:
            thread.join()

        self.flush()

        print(f"Video {self.name} downloaded")

if (__name__ == "__main__"):
    v = Video(name="test/Video_1", url="https://videolectures.unimi.it/vod/mp4:F1X77-23-2023-02-27%2011-06-16.mp4/manifest.m3u8")
    v.download_video()

    print("Exit")