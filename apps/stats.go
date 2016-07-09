package apps

import (
	"fmt"
	"os"
	"time"

	"github.com/Scalingo/cli/config"
	"github.com/Scalingo/go-scalingo"
	tm "github.com/buger/goterm"
	"github.com/jroimartin/gocui"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/errgo.v1"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
)

var (
	stopUI         = make(chan bool)
	stopStats      = make(chan bool)
	refreshUI      = make(chan bool)
	cpuValues      = make(map[string]*tm.DataTable)
	memValues      = make(map[string]*tm.DataTable)
	selectedIndex  = 1
	containerNames = make([]string, 0)
	client         = config.ScalingoClient()
)

func LiveStats(app string) error {
	stats, err := client.AppsStats(app)
	if err != nil {
		return errgo.Mask(err)
	}

	for _, s := range stats.Stats {
		containerNames = append(containerNames, s.ID)
		cpuValues[s.ID] = new(tm.DataTable)
		cpuValues[s.ID].AddColumn("Time")
		cpuValues[s.ID].AddColumn("CPU Usage")
		cpuValues[s.ID].AddRow(float64(time.Now().Unix()), float64(s.CpuUsage))

		memValues[s.ID] = new(tm.DataTable)
		memValues[s.ID].AddColumn("Time")
		memValues[s.ID].AddColumn("RAM")
		memValues[s.ID].AddColumn("SWAP")
		memValues[s.ID].AddRow(float64(time.Now().Unix()), float64(s.MemoryUsage), float64(s.SwapUsage))
	}

	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		return errgo.Mask(err)
	}
	defer g.Close()

	g.SetLayout(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return errgo.Mask(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, previousApp); err != nil {
		return errgo.Mask(err)
	}

	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, nextApp); err != nil {
		return errgo.Mask(err)
	}

	go update(g, app)
	go updateStats(app)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return errgo.Mask(err)
	}

	return nil
}

func nextApp(g *gocui.Gui, v *gocui.View) error {
	if selectedIndex < len(containerNames)-1 {
		selectedIndex++
	}
	refreshUI <- true
	return nil
}

func previousApp(g *gocui.Gui, v *gocui.View) error {
	if selectedIndex > 0 {
		selectedIndex--
	}
	refreshUI <- true
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	stopUI <- true
	stopStats <- true
	return gocui.ErrQuit
}

func update(g *gocui.Gui, app string) {
	g.Execute(func(g *gocui.Gui) error {
		if err := updateView(g); err != nil {
			return errgo.Mask(err)
		}
		return nil
	})

	for {
		select {
		case <-stopUI:
			return
		case <-refreshUI:
			g.Execute(func(g *gocui.Gui) error {
				if err := updateView(g); err != nil {
					return errgo.Mask(err)
				}
				return nil
			})
		}
	}
}

func updateStats(app string) {
	for {
		select {
		case <-stopStats:
			return
		case <-time.After(10 * time.Second):
			stats, err := client.AppsStats(app)
			if err != nil {
				panic(err)
			}
			for _, s := range stats.Stats {
				cpuValues[s.ID].AddRow(float64(time.Now().Unix()), float64(s.CpuUsage))
				memValues[s.ID].AddRow(float64(time.Now().Unix()), float64(s.MemoryUsage), float64(s.SwapUsage))
			}
			refreshUI <- true

		}
	}
}

func updateView(g *gocui.Gui) error {
	if err := updateNames(g); err != nil {
		return errgo.Mask(err)
	}

	if err := updateGraphs(g); err != nil {
		return errgo.Mask(err)
	}
	return nil
}

func updateGraphs(g *gocui.Gui) error {
	v, err := g.View("cpu")
	v.Clear()
	if err != nil {
		return errgo.Mask(err)
	}
	maxX, maxY := v.Size()
	data := cpuValues[containerNames[selectedIndex]]
	chart := tm.NewLineChart(maxX, maxY)
	result := chart.Draw(data)
	fmt.Fprint(v, result)

	v, err = g.View("memory")
	v.Clear()
	if err != nil {
		return errgo.Mask(err)
	}
	maxX, maxY = v.Size()
	data = memValues[containerNames[selectedIndex]]
	chart = tm.NewLineChart(maxX, maxY)
	result = chart.Draw(data)
	fmt.Fprint(v, result)
	return nil
}

func updateNames(g *gocui.Gui) error {
	v, err := g.View("selector")
	if err != nil {
		return errgo.Mask(err)
	}

	maxX, _ := v.Size()
	v.Clear()
	for i := 0; i < maxX/2-5; i++ {
		fmt.Fprint(v, " ")
	}
	fmt.Fprintln(v, "CONTAINERS")
	v.SelBgColor = gocui.ColorGreen
	v.Highlight = true

	for _, name := range containerNames {
		fmt.Fprintln(v, name)
	}
	v.SetCursor(0, selectedIndex+1)

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if _, err := g.SetView("selector", -1, -1, int(0.2*float32(maxX)), maxY); err != nil && err != gocui.ErrUnknownView {
		return errgo.Mask(err)
	}
	if _, err := g.SetView("cpu", int(0.2*float32(maxX)), -1, maxX, maxY/2); err != nil &&
		err != gocui.ErrUnknownView {
		return errgo.Mask(err)
	}

	if _, err := g.SetView("memory", int(0.2*float32(maxX)), maxY/2, maxX, maxY); err != nil &&
		err != gocui.ErrUnknownView {
		return errgo.Mask(err)
	}

	return nil
}

func Stats(app string, stream bool) error {
	if stream {
		return LiveStats(app)
		// c := config.ScalingoClient()
		// stats, err := c.AppsStats(app)
		// if err != nil {
		// 	return errgo.Mask(err)
		// }
		// displayLiveStatsTable(stats.Stats)
		//
		// ticker := time.NewTicker(10 * time.Second)
		// for {
		// 	select {
		// 	case <-ticker.C:
		// 		c := config.ScalingoClient()
		// 		stats, err := c.AppsStats(app)
		// 		if err != nil {
		// 			ticker.Stop()
		// 			return errgo.Mask(err)
		// 		}
		// 		displayLiveStatsTable(stats.Stats)
		// 	}
		// }
	} else {
		c := config.ScalingoClient()
		stats, err := c.AppsStats(app)
		if err != nil {
			return errgo.Mask(err)
		}
		return displayStatsTable(stats.Stats)
	}
}

func displayLiveStatsTable(stats []*scalingo.ContainerStat) {
	fmt.Print("\033[2J\033[;H")
	fmt.Printf("Refreshing every 10 seconds...\n\n")
	displayStatsTable(stats)
	fmt.Println("Last update at:", time.Now().Format(time.UnixDate))
}

func displayStatsTable(stats []*scalingo.ContainerStat) error {
	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Name", "CPU", "Memory", "Swap"})

	for i, s := range stats {
		t.Append([]string{
			s.ID,
			fmt.Sprintf("%d%%", s.CpuUsage),
			fmt.Sprintf(
				"%2d%% %v/%v",
				int(float64(s.MemoryUsage)/float64(s.MemoryLimit)*100),
				toHuman(s.MemoryUsage),
				toHuman(s.MemoryLimit),
			),
			fmt.Sprintf(
				"%2d%% %v/%v",
				int(float64(s.SwapUsage)/float64(s.SwapLimit)*100),
				toHuman(s.SwapUsage),
				toHuman(s.SwapLimit),
			),
		})
		t.Append([]string{
			"", "",
			fmt.Sprintf("Highest: %v", toHuman(s.HighestMemoryUsage)),
			fmt.Sprintf("Highest: %v", toHuman(s.HighestSwapUsage)),
		})
		if i != len(stats)-1 {
			t.Append([]string{"", "", "", ""})
		}
	}

	t.Render()
	return nil
}

func toHuman(sizeInBytes int64) string {
	if sizeInBytes > GB {
		return fmt.Sprintf("%1.1fGB", float64(sizeInBytes)/float64(GB))
	} else if sizeInBytes > MB {
		return fmt.Sprintf("%3dMB", sizeInBytes/MB)
	} else if sizeInBytes > KB {
		return fmt.Sprintf("%3dKB", sizeInBytes/KB)
	} else {
		return fmt.Sprintf("%3dB", sizeInBytes)
	}
}
