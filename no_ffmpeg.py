from dataclasses import dataclass

import imageio.v3 as imageio
import io
import os
import queue
import requests
import threading

import urllib3
urllib3.disable_warnings()

@dataclass
class VideoSegment:
    """VideoSegment contains the data of the video segment downloaded
    from the video manifest and a flag that will tell when the data is
    ready to be written to the file."""

    data: bytes
    """The data of the segment"""

    ready: bool
    """True if the segment is ready to be written"""

    def __init__(self):
        self.data = None
        self.ready = False

@dataclass
class Video:
    """Video is the class that handles the multithreaded download of its
    segments and the creation of the file.

    The download process is regulated by the constants __concurrent_downloads__
    and __flush_after__: the first one determines how many videos can be downloaded
    simultaneously, the second one determines how ofter to write the segments to the
    file. However if the __flush_after__ is less that __concurrent_downloads__, the
    latter will be used"""

    name: str
    """The video name to be used for the file name"""

    url: str
    """The url of the video manifest of type .m3u8"""

    segments: dict[int, VideoSegment]

    ts_file_path: str

    mp4_file_path: str
    
    file: io.BufferedWriter

    download_queue: queue.Queue
    """Queue on which every background thread handling the segment download report
    which segment (by sending its index) is ready to be written. Currently the index
    reported is not used, but clould be for optimization"""

    __concurrent_downloads__ = 8
    """The number of concurrent background threads used to download video segments"""

    __flush_after__ = 20
    """The minimum number of subsequent segments needed before flushing them to the file"""

    def __init__(self, name: str, url: str):
        self.name = name
        self.url = url
        self.segments = dict()
        self.download_queue = queue.Queue()

        video_path = os.getcwd() + "/" + self.name
        self.ts_file_path = video_path + ".ts"
        self.mp4_file_path = video_path + ".mp4"
        # self.file = open(self.ts_file_path, 'wb')

    def download_segment(self, index: int, seg_url: str):
        
        response = requests.get(seg_url, verify=False)

        self.segments[index].data = response.content
        self.segments[index].ready = True

        self.download_queue.put(item=index, block=False)

    def flush(self):
        slice_start = -1

        for i in range(len(self.segments)):
            """Searches the first item that is ready (download completed) and has the data
            field not None (otherwise this means that the segment was already written to the file)"""
            if (self.segments[i].ready and self.segments[i].data != None):
                slice_start = i
                break

        if (slice_start == -1):
            return
        
        slice_end = slice_start
        """Searches how many consecutive segments there are with the same criteria above"""
        for i in range(slice_start+1, len(self.segments)):
            if (self.segments[i].ready == False):
                break
            slice_end = i

        """If there are not enough consecutive segments to be written, it does nothing.
        Flushing too few segments can lead to worse performance and to errors in the
        correct write order (this is unclear why, but if flushing with a rate minor
        to the number of concurrent downloads, the file will be corrupted)"""
        flush_after = max(self.__flush_after__, self.__concurrent_downloads__)
        if (slice_end - slice_start < flush_after - 1):
            return
        
        for i in range(slice_start, slice_end + 1):
            self.file.write(self.segments[i].data)
            """Setting the segment data to None both as a flat to tell that the
            segment was written and to free up the memory"""
            self.segments[i].data = None

    def download_video(self):
        print(f"Downloading video <{self.name}> with url <{self.url}>")
        
        manifest_content = requests.get(self.url, verify=False).text
        video_base_url = self.url[:self.url.rindex("/")]

        segment_manifest_url = video_base_url + "/" + manifest_content.rstrip(" \n").split("\n")[-1]
        segment_manifest_content = requests.get(segment_manifest_url, verify=False).text

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
            """Fills the concurrent background threads with the first
            <__concurrent_downloads__> segments"""
            if (next_downloading < self.__concurrent_downloads__):
                seg_threads[next_downloading].start()
                next_downloading += 1
                continue

            """Waits for one of the segments in download to complete"""
            self.download_queue.get()

            seg_threads[next_downloading].start()
            next_downloading += 1

            self.flush()

        for thread in seg_threads:
            thread.join()

        self.flush()

        print(f"Video <{self.name}> downloaded")

def convert_video(source: str, dest: str):    
    with imageio.imopen(dest, 'w', plugin='pyav') as d:
        source_meta = imageio.immeta(source, plugin='pyav', exclude_applied=False)
        d.init_video_stream(codec=source_meta['codec'], fps=source_meta['fps'])

        for frame in imageio.imiter(source, plugin='pyav'):
            d.write_frame(frame)
        d.close()

if (__name__ == "__main__"):
    v = Video(name="test/Video_1", url="https://videolectures.unimi.it/vod/mp4:F1X77-23-2023-02-27%2011-06-16.mp4/manifest.m3u8")
    # v.download_video()

    convert_video(v.ts_file_path, v.mp4_file_path)

    print("Exit")

