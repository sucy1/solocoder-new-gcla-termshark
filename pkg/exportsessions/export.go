// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package exportsessions

import (
	"bufio"
	"encoding/base64"
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
	StreamIndex int
	Protocol    string
	SrcIP       string
	DstIP       string
	SrcPort     string
	DstPort     string
	Packets     int
	Data        string
}

type SessionExport struct {
	Filename   string
	ExportTime time.Time
	Sessions   []SessionInfo
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
			"-z", fmt.Sprintf("follow,tcp,raw,%d", idx),
		}

		followCmd := exec.Command(tsharkBin, followArgs...)
		followStdout, err := followCmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("creating stdout pipe for follow command stream %d: %w", idx, err)
		}

		if err := followCmd.Start(); err != nil {
			return fmt.Errorf("starting tshark follow command for stream %d: %w", idx, err)
		}

		var reassembledData strings.Builder
		packetCount := 0
		inData := false

		followScanner := bufio.NewScanner(followStdout)
		for followScanner.Scan() {
			line := followScanner.Text()
			if strings.Contains(line, "Follow:") {
				inData = true
				continue
			}
			if inData && strings.TrimSpace(line) == "" {
				continue
			}
			if strings.HasPrefix(line, "===") {
				if inData {
					inData = false
				}
				continue
			}
			if inData {
				trimmed := strings.TrimSpace(line)
				if trimmed != "" {
					reassembledData.WriteString(trimmed)
					packetCount++
				}
			}
		}

		if err := followCmd.Wait(); err != nil {
			return fmt.Errorf("tshark follow command failed for stream %d: %w", idx, err)
		}

		key := streamSet[idx]
		sessions = append(sessions, SessionInfo{
			StreamIndex: idx,
			Protocol:    "TCP",
			SrcIP:       key.srcIP,
			DstIP:       key.dstIP,
			SrcPort:     key.srcPort,
			DstPort:     key.dstPort,
			Packets:     packetCount,
			Data:        base64.StdEncoding.EncodeToString([]byte(reassembledData.String())),
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
