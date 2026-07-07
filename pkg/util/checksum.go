package util

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
)

// ChecksumAlgorithm 校验和算法
type ChecksumAlgorithm string

const (
	AlgorithmMD5    ChecksumAlgorithm = "md5"
	AlgorithmSHA256 ChecksumAlgorithm = "sha256"
)

// CalculateFileChecksum 计算文件校验和
func CalculateFileChecksum(filePath string, algorithm ChecksumAlgorithm) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var hasher hash.Hash
	switch algorithm {
	case AlgorithmMD5:
		hasher = md5.New()
	case AlgorithmSHA256:
		hasher = sha256.New()
	default:
		return "", fmt.Errorf("不支持的算法: %s", algorithm)
	}

	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("计算校验和失败: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// CalculateReaderChecksum 计算 Reader 的校验和
func CalculateReaderChecksum(reader io.Reader, algorithm ChecksumAlgorithm) (string, error) {
	var hasher hash.Hash
	switch algorithm {
	case AlgorithmMD5:
		hasher = md5.New()
	case AlgorithmSHA256:
		hasher = sha256.New()
	default:
		return "", fmt.Errorf("不支持的算法: %s", algorithm)
	}

	if _, err := io.Copy(hasher, reader); err != nil {
		return "", fmt.Errorf("计算校验和失败: %w", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// VerifyFileChecksum 验证文件校验和
func VerifyFileChecksum(filePath string, expectedChecksum string, algorithm ChecksumAlgorithm) (bool, error) {
	actualChecksum, err := CalculateFileChecksum(filePath, algorithm)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}

// ChecksumReader 带校验和计算的 Reader
type ChecksumReader struct {
	reader   io.Reader
	hasher   hash.Hash
	checksum string
}

// NewChecksumReader 创建带校验和的 Reader
func NewChecksumReader(reader io.Reader, algorithm ChecksumAlgorithm) (*ChecksumReader, error) {
	var hasher hash.Hash
	switch algorithm {
	case AlgorithmMD5:
		hasher = md5.New()
	case AlgorithmSHA256:
		hasher = sha256.New()
	default:
		return nil, fmt.Errorf("不支持的算法: %s", algorithm)
	}

	return &ChecksumReader{
		reader: reader,
		hasher: hasher,
	}, nil
}

func (cr *ChecksumReader) Read(p []byte) (int, error) {
	n, err := cr.reader.Read(p)
	if n > 0 {
		cr.hasher.Write(p[:n])
	}
	if err == io.EOF {
		cr.checksum = fmt.Sprintf("%x", cr.hasher.Sum(nil))
	}
	return n, err
}

// Checksum 返回计算的校验和
func (cr *ChecksumReader) Checksum() string {
	return cr.checksum
}

// ChecksumWriter 带校验和计算的 Writer
type ChecksumWriter struct {
	writer   io.Writer
	hasher   hash.Hash
	checksum string
}

// NewChecksumWriter 创建带校验和的 Writer
func NewChecksumWriter(writer io.Writer, algorithm ChecksumAlgorithm) (*ChecksumWriter, error) {
	var hasher hash.Hash
	switch algorithm {
	case AlgorithmMD5:
		hasher = md5.New()
	case AlgorithmSHA256:
		hasher = sha256.New()
	default:
		return nil, fmt.Errorf("不支持的算法: %s", algorithm)
	}

	return &ChecksumWriter{
		writer: writer,
		hasher: hasher,
	}, nil
}

func (cw *ChecksumWriter) Write(p []byte) (int, error) {
	n, err := cw.writer.Write(p)
	if n > 0 {
		cw.hasher.Write(p[:n])
	}
	return n, err
}

// Checksum 返回计算的校验和
func (cw *ChecksumWriter) Checksum() string {
	if cw.checksum == "" {
		cw.checksum = fmt.Sprintf("%x", cw.hasher.Sum(nil))
	}
	return cw.checksum
}
