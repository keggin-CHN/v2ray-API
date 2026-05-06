package xrayruntime

import "strings"

func NormalizeParsedNode(node ParsedNode) ParsedNode {
	if node.Network == "" {
		node.Network = "tcp"
	}
	if node.Path == "" && node.Network == "ws" {
		node.Path = "/"
	}
	if node.Scheme == "trojan" && node.Security == "" {
		node.Security = "tls"
	}
	if node.Scheme == "hy2" || node.Scheme == "hysteria2" {
		node.Scheme = "hysteria2"
		if node.Security == "" {
			node.Security = "tls"
		}
	}
	if node.Scheme == "https" {
		node.Scheme = "http"
	}
	if node.SNI == "" {
		node.SNI = node.HostHdr
	}
	node.Name = strings.TrimSpace(node.Name)
	return node
}
