// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package exportsessions

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/gcla/termshark/v2"
)

type SessionInfo struct {
	StreamIndex int    `json:"stream_index"`
	Protocol    string `json:"protocol"`
	SrcIP       string `json:"src_ip"`
	DstIP       string `json:"dst_ip"`
	SrcPort     string `json:"src_port"`
	DstPort     string `json:"dst_port"`
	Packets     int    `json:"packets"`
	Data        string `json:"data"`
}

type SessionExport struct {
	Filename   string        `json:"filename"`
	ExportTime time.Time     `json:"export_time"`
	Sessions   []SessionInfo `json:"sessions"`
}

func parseHexFollowOutput(text string) []byte {
	var buf []byte
	inData := false
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Follow:") {
			inData = true
			continue
		}
		if !inData {
			continue
		}
		if strings.HasPrefix(trimmed, "===") {
			inData = false
			continue
		}
		if trimmed == "" {
			continue
		}
		hexPart := trimmed
		if idx := strings.Index(trimmed, "  "); idx > 0 {
			hexPart = trimmed[:idx]
		}
		hexPart = strings.ReplaceAll(hexPart, " ", "")
		decoded, err := hex.DecodeString(hexPart)
		if err != nil {
			continue
		}
		buf = append(buf, decoded...)
	}
	return buf
}

func ExportSessions(pcapFile string, tsharkBin string, outputFile string) error {
	if tsharkBin == "" {
		tsharkBin = termshark.TSharkBin()
	}

	args := []string{
		"-r", pcapFile,
		"-T", "fields",
		"-e", "tcp.stream",
		"-e", "ip.src",
		"-e", "ip.dst",
		"-e", "tcp.srcport",
		"-e", "tcp.dstport",
	}

	cmd := exec.Command(tsharkBin, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting tshark fields command: %w", err)
	}

	type streamKey struct {
		index   int
		srcIP   string
		dstIP   string
		srcPort string
		dstPort string
	}

	streamSet := make(map[int]streamKey)
	packetCountMap := make(map[int]int)
	uniqueIndices := make(map[int]bool)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(fields[0], "%d", &idx); err != nil {
			continue
		}
		uniqueIndices[idx] = true
		packetCountMap[idx]++
		if _, exists := streamSet[idx]; !exists {
			streamSet[idx] = streamKey{
				index:   idx,
				srcIP:   fields[1],
				dstIP:   fields[2],
				srcPort: fields[3],
				dstPort: fields[4],
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("tshark fields command failed: %w", err)
	}

	sortedIndices := make([]int, 0, len(uniqueIndices))
	for idx := range uniqueIndices {
		sortedIndices = append(sortedIndices, idx)
	}
	sort.Ints(sortedIndices)

	sessions := make([]SessionInfo, 0, len(sortedIndices))

	for _, idx := range sortedIndices {
		followArgs := []string{
			"-r", pcapFile,
			"-q",
			"-z", fmt.Sprintf("follow,tcp,hex,%d", idx),
		}

		var stdoutBuf strings.Builder
		followCmd := exec.Command(tsharkBin, followArgs...)
		followCmd.Stdout = &stdoutBuf
		followCmd.Stderr = os.Stderr

		if err := followCmd.Run(); err != nil {
			return fmt.Errorf("tshark follow command failed for stream %d: %w", idx, err)
		}

		binaryData := parseHexFollowOutput(stdoutBuf.String())
		b64Data := base64.StdEncoding.EncodeToString(binaryData)

		key := streamSet[idx]
		pktCount := packetCountMap[idx]
		if pktCount == 0 {
			pktCount = 1
		}
		sessions = append(sessions, SessionInfo{
			StreamIndex: idx,
			Protocol:    "TCP",
			SrcIP:       key.srcIP,
			DstIP:       key.dstIP,
			SrcPort:     key.srcPort,
			DstPort:     key.dstPort,
			Packets:     pktCount,
			Data:        b64Data,
		})
	}

	export := SessionExport{
		Filename:   pcapFile,
		ExportTime: time.Now(),
		Sessions:   sessions,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}
