package build

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type Vulnerability struct {
	ID          string `json:"id"`
	Package     string `json:"package"`
	Version     string `json:"version"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

type ScanResult struct {
	TotalVulnerabilities int             `json:"total_vulnerabilities"`
	Critical             int             `json:"critical"`
	High                 int             `json:"high"`
	Medium               int             `json:"medium"`
	Low                  int             `json:"low"`
	Vulnerabilities      []Vulnerability `json:"vulnerabilities"`
}

type trivyOutput struct {
	Results []trivyResult `json:"Results"`
}

type trivyResult struct {
	Vulnerabilities []trivyVuln `json:"Vulnerabilities"`
}

type trivyVuln struct {
	VulnerabilityID  string `json:"VulnerabilityID"`
	PkgName          string `json:"PkgName"`
	InstalledVersion string `json:"InstalledVersion"`
	Severity         string `json:"Severity"`
	Description      string `json:"Description"`
}

func ScanImage(imageRef string) (*ScanResult, error) {
	cmd := exec.Command("trivy", "image", "--format", "json", imageRef)
	output, err := cmd.Output()
	if err != nil {
		// Trivy returns non-zero exit code when vulnerabilities are found,
		// but we still want to parse the output
		if exitErr, ok := err.(*exec.ExitError); ok && len(output) == 0 {
			return nil, fmt.Errorf("trivy scan failed: %s", string(exitErr.Stderr))
		}
	}

	var trivyOut trivyOutput
	if err := json.Unmarshal(output, &trivyOut); err != nil {
		return nil, fmt.Errorf("failed to parse trivy output: %w", err)
	}

	result := &ScanResult{}

	for _, r := range trivyOut.Results {
		for _, v := range r.Vulnerabilities {
			vuln := Vulnerability{
				ID:          v.VulnerabilityID,
				Package:     v.PkgName,
				Version:     v.InstalledVersion,
				Severity:    v.Severity,
				Description: v.Description,
			}
			result.Vulnerabilities = append(result.Vulnerabilities, vuln)
			result.TotalVulnerabilities++

			switch v.Severity {
			case "CRITICAL":
				result.Critical++
			case "HIGH":
				result.High++
			case "MEDIUM":
				result.Medium++
			case "LOW":
				result.Low++
			}
		}
	}

	return result, nil
}
