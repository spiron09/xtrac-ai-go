package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/k3a/html2text"
	"google.golang.org/api/gmail/v1"
)

func getHeaderValue(headers []*gmail.MessagePartHeader, name string) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}

func decodeBase64(data string) (string, error) {
	// Replace URL-safe characters and add padding if needed
	data = strings.Replace(data, "-", "+", -1)
	data = strings.Replace(data, "_", "/", -1)
	
	// Add padding if necessary
	for len(data)%4 != 0 {
		data += "="
	}
	
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Helper function to recursively extract content from message parts
func extractContentFromPart(part *gmail.MessagePart) (string, error) {
	if part == nil {
		return "", nil
	}
	
	// If this part has data, decode it
	if part.Body != nil && part.Body.Data != "" {
		decoded, err := decodeBase64(part.Body.Data)
		if err != nil {
			return "", err
		}
		
		// Convert HTML to plain text if needed
		if strings.HasPrefix(part.MimeType, "text/html") {
			text := html2text.HTML2Text(decoded)
			return text, nil
		}
		
		return decoded, nil
	}
	
	// If this part has nested parts, recursively process them
	if len(part.Parts) > 0 {
		var content strings.Builder
		for _, subPart := range part.Parts {
			subContent, err := extractContentFromPart(subPart)
			if err != nil {
				return "", err
			}
			content.WriteString(subContent)
		}
		return content.String(), nil
	}
	
	return "", nil
}

// Find the first part with a specific MIME type
func findPartByMimeType(parts []*gmail.MessagePart, mimeType string) *gmail.MessagePart {
	for _, part := range parts {
		if part.MimeType == mimeType {
			return part
		}
		// Recursively check nested parts
		if len(part.Parts) > 0 {
			if found := findPartByMimeType(part.Parts, mimeType); found != nil {
				return found
			}
		}
	}
	return nil
}

func parseTextPlain(m *gmail.Message) (string, error) {
	fmt.Println("Parsing text/plain")
	p := m.Payload
	if p == nil || p.Body == nil {
		fmt.Println("text/plain: no payload/body")
		return "", nil
	}
	
	var body string
	if p.Body.Data != "" {
		decoded, err := decodeBase64(p.Body.Data)
		if err != nil {
			return "", err
		}
		body = decoded
	}
	return body, nil
}

func parseTextHtml(m *gmail.Message) (string, error) {
	fmt.Println("Parsing text/html")
	p := m.Payload
	if p == nil || p.Body == nil {
		fmt.Println("text/html: no payload/body")
		return "", nil
	}
	
	var body string
	if p.Body.Data != "" {
		decoded, err := decodeBase64(p.Body.Data)
		if err != nil {
			return "", err
		}
		// Convert HTML to plain text
		text := html2text.HTML2Text(decoded)
		body = text
	}
	return body, nil
}

func parseMultipartAlternative(m *gmail.Message) (string, error) {
	fmt.Println("Parsing multipart/alternative")
	p := m.Payload
	if p == nil {
		fmt.Println("multipart/alternative: no payload")
		return "", nil
	}
	
	// multipart/alternative typically contains the same content in different formats
	// Prefer text/plain over text/html if available
	if textPart := findPartByMimeType(p.Parts, "text/plain"); textPart != nil {
		return extractContentFromPart(textPart)
	}
	
	// Fall back to text/html if text/plain is not available
	if htmlPart := findPartByMimeType(p.Parts, "text/html"); htmlPart != nil {
		return extractContentFromPart(htmlPart)
	}
	
	// If no preferred parts found, extract content from the first available part
	if len(p.Parts) > 0 {
		return extractContentFromPart(p.Parts[0])
	}
	
	fmt.Println("multipart/alternative: no parts found")
	return "", nil
}

func parseMultipartMixed(m *gmail.Message) (string, error) {
	fmt.Println("Parsing multipart/mixed")
	p := m.Payload
	if p == nil {
		fmt.Println("multipart/mixed: no payload")
		return "", nil
	}
	
	var content strings.Builder
	
	// multipart/mixed can contain various parts including attachments
	// Extract content from text parts
	for _, part := range p.Parts {
		if strings.HasPrefix(part.MimeType, "text/") {
			partContent, err := extractContentFromPart(part)
			if err != nil {
				return "", err
			}
			content.WriteString(partContent)
		}
	}
	
	// If no text parts were found, try to extract from any part
	if content.Len() == 0 && len(p.Parts) > 0 {
		for _, part := range p.Parts {
			partContent, err := extractContentFromPart(part)
			if err != nil {
				return "", err
			}
			content.WriteString(partContent)
			if partContent != "" {
				break // Stop after finding the first non-empty part
			}
		}
	}
	
	return content.String(), nil
}