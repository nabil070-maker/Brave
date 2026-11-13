package main

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

//go:embed assets/*.png
var logos embed.FS

// set at build time: -X main.channel=stable|beta|nightly
var channel = "stable"

type chInfo struct{ Title, Prefix string }

var channels = map[string]chInfo{
	"stable":  {"Brave", "Release"},
	"beta":    {"Brave Beta", "Beta"},
	"nightly": {"Brave Nightly", "Nightly"},
}

type release struct {
	Name   string `json:"name"`
	Draft  bool   `json:"draft"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func latest() (name, url string, err error) {
	resp, err := http.Get("https://api.github.com/repos/brave/brave-browser/releases?per_page=100")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var rs []release
	if err = json.NewDecoder(resp.Body).Decode(&rs); err != nil {
		return
	}
	prefix := channels[channel].Prefix
	for _, r := range rs {
		if r.Draft || !strings.HasPrefix(r.Name, prefix) {
			continue
		}
		for _, a := range r.Assets {
			if strings.HasPrefix(a.Name, "brave-") && strings.HasSuffix(a.Name, "win32-x64.zip") {
				return r.Name, a.URL, nil
			}
		}
	}
	return "", "", fmt.Errorf("no %s build found", channel)
}

type progressWriter struct {
	total, done int64
	onUpdate    func(int)
}

func (p *progressWriter) Write(b []byte) (int, error) {
	p.done += int64(len(b))
	if p.total > 0 {
		p.onUpdate(int(p.done * 100 / p.total))
	}
	return len(b), nil
}

func install(url, appDir string, onProgress func(int)) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	pw := &progressWriter{total: resp.ContentLength, onUpdate: onProgress}
	if _, err = io.Copy(&buf, io.TeeReader(resp.Body, pw)); err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		dest := filepath.Join(appDir, f.Name)
		if !strings.HasPrefix(dest, filepath.Clean(appDir)+string(os.PathSeparator)) {
			continue // zip-slip guard
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(dest, 0o755)
			continue
		}
		os.MkdirAll(filepath.Dir(dest), 0o755)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func launch(root string) {
	profile := filepath.Join(root, "data", "profile")
	os.MkdirAll(profile, 0o755)
	cmd := exec.Command(filepath.Join(root, "app", "brave.exe"),
		"--user-data-dir="+profile,
		"--no-default-browser-check",
		"--disable-breaking-changes")
	cmd.Start()
}

func main() {
	exe, _ := os.Executable()
	root := filepath.Dir(exe)
	appDir := filepath.Join(root, "app")

	// already installed -> just launch
	if _, err := os.Stat(filepath.Join(appDir, "brave.exe")); err == nil {
		launch(root)
		return
	}

	info := channels[channel]
	data, _ := logos.ReadFile("assets/" + channel + ".png")
	img, _ := png.Decode(bytes.NewReader(data))
	bmp, _ := walk.NewBitmapFromImage(img)

	var mw *walk.MainWindow
	var status *walk.Label
	var pb *walk.ProgressBar
	var btn *walk.PushButton
	done := false

	setStatus := func(s string) { mw.Synchronize(func() { status.SetText(s) }) }
	enable := func() { mw.Synchronize(func() { btn.SetEnabled(true) }) }

	MainWindow{
		AssignTo: &mw,
		Title:    info.Title + " Portable Setup",
		MinSize:  Size{Width: 460, Height: 340},
		Size:     Size{Width: 460, Height: 340},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}},
		Children: []Widget{
			ImageView{Image: bmp, Mode: ImageViewModeZoom, MinSize: Size{Width: 96, Height: 96}, MaxSize: Size{Width: 96, Height: 96}},
			Label{Text: info.Title + " Portable", Font: Font{PointSize: 16, Bold: true}, TextAlignment: AlignCenter},
			Label{AssignTo: &status, Text: "Ready to install into the 'app' folder next to this file.", TextAlignment: AlignCenter},
			ProgressBar{AssignTo: &pb, MinValue: 0, MaxValue: 100},
			PushButton{
				AssignTo: &btn,
				Text:     "Install",
				OnClicked: func() {
					if done {
						launch(root)
						mw.Close()
						return
					}
					btn.SetEnabled(false)
					go func() {
						setStatus("Looking up the latest " + info.Title + "...")
						name, url, err := latest()
						if err != nil {
							setStatus("Error: " + err.Error())
							enable()
							return
						}
						setStatus("Downloading " + name + "...")
						err = install(url, appDir, func(p int) {
							mw.Synchronize(func() { pb.SetValue(p) })
						})
						if err != nil {
							setStatus("Error: " + err.Error())
							enable()
							return
						}
						done = true
						setStatus("Installed " + name)
						mw.Synchronize(func() {
							btn.SetText("Launch")
							btn.SetEnabled(true)
						})
					}()
				},
			},
		},
	}.Run()
}
