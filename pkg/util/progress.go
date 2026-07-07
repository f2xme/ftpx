package util

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v3"
)

// ProgressBar 进度条包装器
type ProgressBar struct {
	bar       *progressbar.ProgressBar
	total     int64
	startTime time.Time
}

// NewProgressBar 创建进度条
func NewProgressBar(total int64, description string) *ProgressBar {
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("B/s"),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			fmt.Println()
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

	return &ProgressBar{
		bar:       bar,
		total:     total,
		startTime: time.Now(),
	}
}

// Update 更新进度
func (pb *ProgressBar) Update(current int64) error {
	return pb.bar.Set64(current)
}

// Add 增加进度
func (pb *ProgressBar) Add(delta int) error {
	return pb.bar.Add(delta)
}

// Finish 完成进度条
func (pb *ProgressBar) Finish() error {
	return pb.bar.Finish()
}

// Clear 清除进度条
func (pb *ProgressBar) Clear() error {
	return pb.bar.Clear()
}

// GetElapsed 获取经过的时间
func (pb *ProgressBar) GetElapsed() time.Duration {
	return time.Since(pb.startTime)
}

// GetSpeed 获取当前速度（字节/秒）
func (pb *ProgressBar) GetSpeed(current int64) float64 {
	elapsed := time.Since(pb.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(current) / elapsed
}

// FormatBytes 格式化字节数
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), units[exp])
}

// FormatSpeed 格式化速度
func FormatSpeed(bytesPerSecond float64) string {
	return FormatBytes(int64(bytesPerSecond)) + "/s"
}

// FormatDuration 格式化时间
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

// SimpleProgressBar 简单进度条（用于不支持 progressbar 的场景）
type SimpleProgressBar struct {
	total       int64
	current     int64
	startTime   time.Time
	description string
}

// NewSimpleProgressBar 创建简单进度条
func NewSimpleProgressBar(total int64, description string) *SimpleProgressBar {
	return &SimpleProgressBar{
		total:       total,
		current:     0,
		startTime:   time.Now(),
		description: description,
	}
}

// Update 更新进度
func (spb *SimpleProgressBar) Update(current int64) error {
	spb.current = current
	return spb.render()
}

// Add 增加进度
func (spb *SimpleProgressBar) Add(delta int) error {
	spb.current += int64(delta)
	return spb.render()
}

// render 渲染进度条
func (spb *SimpleProgressBar) render() error {
	percent := float64(spb.current) / float64(spb.total) * 100
	elapsed := time.Since(spb.startTime).Seconds()
	speed := float64(spb.current) / elapsed

	fmt.Printf("\r%s: %.1f%% (%s/%s) @ %s",
		spb.description,
		percent,
		FormatBytes(spb.current),
		FormatBytes(spb.total),
		FormatSpeed(speed),
	)

	return nil
}

// Finish 完成进度条
func (spb *SimpleProgressBar) Finish() error {
	fmt.Println()
	return nil
}

// Clear 清除进度条
func (spb *SimpleProgressBar) Clear() error {
	fmt.Print("\r" + string(make([]byte, 80)) + "\r")
	return nil
}
