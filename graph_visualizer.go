package main

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strings"

	"github.com/fogleman/gg"
)

// Node represents a page node in the graph
type Node struct {
	URL      string
	X        float64
	Y        float64
	Radius   float64
	Color    [3]float64 // RGB values
	IsExternal bool
}

// Edge represents a link between pages
type Edge struct {
	From string
	To   string
	Weight int
}

// GraphVisualizer handles the creation of graph visualizations
type GraphVisualizer struct {
	nodes map[string]*Node
	edges []Edge
	width int
	height int
}

// NewGraphVisualizer creates a new graph visualizer
func NewGraphVisualizer(width, height int) *GraphVisualizer {
	return &GraphVisualizer{
		nodes:  make(map[string]*Node),
		edges:  make([]Edge, 0),
		width:  width,
		height: height,
	}
}

// AddInternalPages adds internal pages to the graph
func (gv *GraphVisualizer) AddInternalPages(pages map[string]int, baseURL string) {
	// Parse base URL to get domain
	parsedBase, _ := url.Parse(baseURL)
	
	// Convert pages to sorted slice for consistent positioning
	type PageInfo struct {
		URL   string
		Count int
	}
	var pageList []PageInfo
	for normalizedURL, count := range pages {
		// Reconstruct full URL
		fullURL := parsedBase.Scheme + "://" + normalizedURL
		pageList = append(pageList, PageInfo{URL: fullURL, Count: count})
	}
	
	// Sort by count (descending) for better visualization
	sort.Slice(pageList, func(i, j int) bool {
		return pageList[i].Count > pageList[j].Count
	})
	
	// Position nodes in a circle for internal pages
	centerX := float64(gv.width) * 0.3
	centerY := float64(gv.height) * 0.5
	radius := math.Min(float64(gv.width), float64(gv.height)) * 0.2
	
	for i, page := range pageList {
		angle := 2 * math.Pi * float64(i) / float64(len(pageList))
		x := centerX + radius*math.Cos(angle)
		y := centerY + radius*math.Sin(angle)
		
		// Node size based on link count
		nodeRadius := 5 + float64(page.Count)*2
		if nodeRadius > 20 {
			nodeRadius = 20
		}
		
		gv.nodes[page.URL] = &Node{
			URL:      page.URL,
			X:        x,
			Y:        y,
			Radius:   nodeRadius,
			Color:    [3]float64{0.2, 0.6, 0.9}, // Blue for internal
			IsExternal: false,
		}
	}
}

// AddExternalLinks adds external links to the graph
func (gv *GraphVisualizer) AddExternalLinks(externalLinks map[string]int) {
	// Convert to sorted slice
	type ExternalInfo struct {
		URL   string
		Count int
	}
	var extList []ExternalInfo
	for url, count := range externalLinks {
		extList = append(extList, ExternalInfo{URL: url, Count: count})
	}
	
	// Sort by count (descending)
	sort.Slice(extList, func(i, j int) bool {
		return extList[i].Count > extList[j].Count
	})
	
	// Position external nodes on the right side
	startX := float64(gv.width) * 0.7
	startY := float64(gv.height) * 0.1
	spacing := float64(gv.height) * 0.8 / float64(len(extList)+1)
	
	for i, ext := range extList {
		y := startY + float64(i+1)*spacing
		
		// Node size based on link count
		nodeRadius := 3 + float64(ext.Count)*1.5
		if nodeRadius > 15 {
			nodeRadius = 15
		}
		
		gv.nodes[ext.URL] = &Node{
			URL:      ext.URL,
			X:        startX,
			Y:        y,
			Radius:   nodeRadius,
			Color:    [3]float64{0.9, 0.4, 0.2}, // Orange for external
			IsExternal: true,
		}
	}
}

// AddEdges creates edges between nodes based on discovered links
func (gv *GraphVisualizer) AddEdges(pages map[string]int, externalLinks map[string]int, baseURL string) {
	// Parse base URL
	parsedBase, _ := url.Parse(baseURL)
	
	// Create edges between internal pages (simplified - all connected to main page)
	mainURL := baseURL
	for normalizedURL := range pages {
		fullURL := parsedBase.Scheme + "://" + normalizedURL
		if fullURL != mainURL {
			gv.edges = append(gv.edges, Edge{
				From:   mainURL,
				To:     fullURL,
				Weight: pages[normalizedURL],
			})
		}
	}
	
	// Create edges to external links (from main page)
	for extURL, count := range externalLinks {
		gv.edges = append(gv.edges, Edge{
			From:   mainURL,
			To:     extURL,
			Weight: count,
		})
	}
}

// DrawGraph creates the visualization and saves it to a file
func (gv *GraphVisualizer) DrawGraph(filename string) error {
	dc := gg.NewContext(gv.width, gv.height)
	
	// Set background
	dc.SetRGB(1, 1, 1) // White background
	dc.Clear()
	
	// Draw edges first (so they appear behind nodes)
	for _, edge := range gv.edges {
		fromNode := gv.nodes[edge.From]
		toNode := gv.nodes[edge.To]
		
		if fromNode != nil && toNode != nil {
			// Line thickness based on weight
			lineWidth := 1.0 + float64(edge.Weight)*0.5
			if lineWidth > 5 {
				lineWidth = 5
			}
			
			// Different colors for internal vs external edges
			if toNode.IsExternal {
				dc.SetRGB(0.9, 0.4, 0.2) // Orange for external links
			} else {
				dc.SetRGB(0.2, 0.6, 0.9) // Blue for internal links
			}
			
			dc.SetLineWidth(lineWidth)
			dc.DrawLine(fromNode.X, fromNode.Y, toNode.X, toNode.Y)
			dc.Stroke()
		}
	}
	
	// Draw nodes
	for _, node := range gv.nodes {
		// Draw node circle
		dc.SetRGB(node.Color[0], node.Color[1], node.Color[2])
		dc.DrawCircle(node.X, node.Y, node.Radius)
		dc.Fill()
		
		// Draw node border
		dc.SetRGB(0, 0, 0)
		dc.SetLineWidth(1)
		dc.DrawCircle(node.X, node.Y, node.Radius)
		dc.Stroke()
	}
	
	// Draw labels for nodes
	dc.SetRGB(0, 0, 0)
	fontSize := 10.0
	if err := dc.LoadFontFace("/System/Library/Fonts/Arial.ttf", fontSize); err != nil {
		// Try alternative system fonts
		if err := dc.LoadFontFace("/System/Library/Fonts/Helvetica.ttc", fontSize); err != nil {
			// If no system fonts work, just use smaller text without custom font
			fontSize = 12.0
		}
	}
	
	for _, node := range gv.nodes {
		// Create short label from URL
		label := gv.createShortLabel(node.URL)
		
		// Position label slightly offset from node
		labelX := node.X + node.Radius + 5
		labelY := node.Y + 3
		
		// Ensure label doesn't go off screen
		if labelX > float64(gv.width)-100 {
			labelX = node.X - node.Radius - 5
		}
		
		dc.DrawString(label, labelX, labelY)
	}
	
	// Add title
	dc.SetRGB(0, 0, 0)
	titleSize := 16.0
	if err := dc.LoadFontFace("/System/Library/Fonts/Arial.ttf", titleSize); err != nil {
		if err := dc.LoadFontFace("/System/Library/Fonts/Helvetica.ttc", titleSize); err != nil {
			titleSize = 18.0
		}
	}
	dc.DrawString("Web Crawler Link Graph", 20, 30)
	
	// Add legend
	dc.SetRGB(0, 0, 0)
	legendSize := 12.0
	if err := dc.LoadFontFace("/System/Library/Fonts/Arial.ttf", legendSize); err != nil {
		if err := dc.LoadFontFace("/System/Library/Fonts/Helvetica.ttc", legendSize); err != nil {
			legendSize = 14.0
		}
	}
	
	legendY := float64(gv.height) - 60
	
	// Internal links legend
	dc.SetRGB(0.2, 0.6, 0.9)
	dc.DrawCircle(20, legendY, 8)
	dc.Fill()
	dc.SetRGB(0, 0, 0)
	dc.DrawString("Internal Pages", 35, legendY+4)
	
	// External links legend
	dc.SetRGB(0.9, 0.4, 0.2)
	dc.DrawCircle(20, legendY+20, 8)
	dc.Fill()
	dc.SetRGB(0, 0, 0)
	dc.DrawString("External Links", 35, legendY+24)
	
	// Save the image
	return dc.SavePNG(filename)
}

// createShortLabel creates a short, readable label from a URL
func (gv *GraphVisualizer) createShortLabel(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	
	// For external links, show domain
	if strings.Contains(parsed.Host, ".") {
		if parsed.Path == "" || parsed.Path == "/" {
			return parsed.Host
		}
		// Show domain + first path segment
		pathParts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(pathParts) > 0 && pathParts[0] != "" {
			return parsed.Host + "/" + pathParts[0]
		}
		return parsed.Host
	}
	
	// For internal links, show path
	if parsed.Path == "" || parsed.Path == "/" {
		return "/"
	}
	
	pathParts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(pathParts) > 2 {
		return "/" + pathParts[0] + "/..."
	}
	return parsed.Path
}

// GenerateGraphVisualization creates a complete graph visualization
func GenerateGraphVisualization(pages map[string]int, externalLinks map[string]int, baseURL, filename string) error {
	// Create visualizer
	gv := NewGraphVisualizer(1200, 800)
	
	// Add data to graph
	gv.AddInternalPages(pages, baseURL)
	gv.AddExternalLinks(externalLinks)
	gv.AddEdges(pages, externalLinks, baseURL)
	
	// Generate the image
	if err := gv.DrawGraph(filename); err != nil {
		return fmt.Errorf("failed to generate graph: %v", err)
	}
	
	fmt.Printf("Graph visualization saved to: %s\n", filename)
	return nil
}
