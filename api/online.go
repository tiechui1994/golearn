package api

import (
	"time"
	"net/http"
	"net"
	"crypto/tls"
	"net/url"
	"os"
	"fmt"
	"strings"
	"log"
	"bytes"
	"mime/multipart"
	"io/ioutil"
	"encoding/json"
)

//https://s113.123apps.com/aconv/upload/flow/?uid=2VlBGJGDkxWMXGYGHs65e5cba075f488&
// id3=true&
// ff=true&
// flowChunkNumber=1&
// flowChunkSize=52428800&
// flowCurrentChunkSize=6182&
// flowTotalSize=6182&
// flowIdentifier=6182-Light_Google_002_set50_enamr&
// flowFilename=Light_Google_002_set50_en.amr&
// flowRelativePath=Light_Google_002_set50_en.amr&
// flowTotalChunks=1

var oclient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Second*10)
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

const (
	uid           = "2VlBGJGDkxWMXGYGHs65e5cba075f488"
	flowChunkSize = 52428800
)

type audio struct {
	Index          int    `json:"index"`
	CodecName      string `json:"codec_name"`
	CodecLongName  string `json:"codec_long_name"`
	CodecType      string `json:"codec_type"`
	CodecTimeBase  string `json:"codec_time_base"`
	CodecTagString string `json:"codec_tag_string"`
	CodecTag       string `json:"codec_tag"`
	SampleFmt      string `json:"sample_fmt"`
	SampleRate     string `json:"sample_rate"`
	Channels       int    `json:"channels"`
	ChannelLayout  string `json:"channel_layout"`
	BitsPerSample  int    `json:"bits_per_sample"`
	RFrameRate     string `json:"r_frame_rate"`
	AvgFrameRate   string `json:"avg_frame_rate"`
	TimeBase       string `json:"time_base"`
	StartPts       int    `json:"start_pts"`
	StartTime      string `json:"start_time"`
	DurationTs     int64  `json:"duration_ts"`
	Duration       string `json:"duration"`
	BitRate        string `json:"bit_rate"`
	Disposition struct {
		Default         int `json:"default"`
		Dub             int `json:"dub"`
		Original        int `json:"original"`
		Comment         int `json:"comment"`
		Lyrics          int `json:"lyrics"`
		Karaoke         int `json:"karaoke"`
		Forced          int `json:"forced"`
		HearingImpaired int `json:"hearing_impaired"`
		VisualImpaired  int `json:"visual_impaired"`
		CleanEffects    int `json:"clean_effects"`
		AttachedPic     int `json:"attached_pic"`
		TimedThumbnails int `json:"timed_thumbnails"`
	} `json:"disposition"`
}

func UploadFlow(filename string) error {
	fd, _ := os.Open(filename)
	stat, _ := os.Stat(filename)
	n := stat.Size() / flowChunkSize
	if stat.Size()%flowChunkSize != 0 {
		n += 1
	}

	data := make([]byte, flowChunkSize)
	total := int64(0)
	for i := int64(1); i <= n; i++ {
		currentSize := int64(flowChunkSize)
		if currentSize >= stat.Size()-total {
			currentSize = stat.Size() - total
		}

		values := make(url.Values)
		values.Set("uid", uid)
		values.Set("id3", "true")
		values.Set("ff", "true")
		values.Set("flowChunkNumber", fmt.Sprintf("%v", i))
		values.Set("flowChunkSize", fmt.Sprintf("%v", flowChunkSize))
		values.Set("flowCurrentChunkSize", fmt.Sprintf("%v", currentSize))
		values.Set("flowTotalSize", fmt.Sprintf("%v", stat.Size()))
		values.Set("flowIdentifier", fmt.Sprintf("%v-%v", stat.Size(), strings.Replace(stat.Name(), ".", "", -1)))
		values.Set("flowFilename", stat.Name())
		values.Set("flowRelativePath", stat.Name())
		values.Set("flowTotalChunks", fmt.Sprintf("%v", n))

		n, _ := fd.ReadAt(data, total)
		data = data[:n]
		total += flowChunkSize

		u := "https://s113.123apps.com/aconv/upload/flow/?" + values.Encode()
		request, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return err
		}

		response, err := oclient.Do(request)
		if err != nil {
			return err
		}

		log.Println("Status:", response.StatusCode)

		Flow(values, data)
	}

	return nil
}

func Flow(vals url.Values, data []byte) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for k, v := range vals {
		writer.WriteField(k, v[0])
	}

	fd, err := writer.CreateFormFile("file", vals.Get("flowFilename"))
	if err != nil {
		return err
	}
	fd.Write(data)

	contentType := writer.FormDataContentType()
	writer.Close()

	u := "https://s113.123apps.com/aconv/upload/flow/"
	request, err := http.NewRequest("POST", u, &body)
	request.Header.Set("Content-Type", contentType)

	response, err := oclient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Printf("%v", string(bs))
	var result struct {
		Error            int    `json:"error"`
		FileSize         int64  `json:"filesize"`
		TmpFilename      string `json:"tmp_filename"`
		OriginalFilename string `json:"original_filename"`
		FileType         string `json:"file_type"`
		Id3 struct {
			TagWasRead          bool   `json:"tag_was_read"`
			TrackTagGenre       string `json:"track_tag_genre"`
			TrackTagTracknumber string `json:"track_tag_tracknumber"`
			Bitrate             int    `json:"bitrate"`
			TrackBitrate        int    `json:"track_bitrate"`
			DurationInSeconds   int    `json:"duration_in_seconds"`
		} `json:"id3"`
		Ff struct {
			FfprobeSuccess         bool `json:"ffprobe_success"`
			DurationInSeconds      int  `json:"duration_in_seconds"`
			DurationInMilliseconds int  `json:"duration_in_milliseconds"`
			Bitrate                int  `json:"bitrate"`
			HasAudioStreams        bool `json:"has_audio_streams"`
			HasVedioStreams        bool `json:"has_vedio_streams"`
			Filesize               int  `json:"filesize"`
			Streams struct {
				Audio []audio `json:"audio"`
			} `json:"streams"`
			Format struct {
				NbStreams      int    `json:"nb_streams"`
				NbPrograms     int    `json:"nb_programs"`
				FormatName     string `json:"format_name"`
				FormatLongName string `json:"format_long_name"`
				Duration       string `json:"duration"`
				BitRate        string `json:"bit_rate"`
				ProbeScore     int    `json:"probe_score"`
			} `json:"format"`
		} `json:"ff"`
	}
	err = json.Unmarshal(bs, &result)
	if err != nil {
		return err
	}

	log.Printf("result: %+v", result)
	return nil
}
