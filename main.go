package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/olekukonko/tablewriter"
	"github.com/samjohnduke/sUPnP"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("help"),
	readline.PcItem("list"),
	readline.PcItem("ip"),
	readline.PcItem("map",
		readline.PcItem("add"),
		readline.PcItem("remove"),
	),
	readline.PcItem("exit"),
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func main() {
	client, err := sUPnP.Discover()
	if err != nil {
		//No available UPnP router, not much we can do but bail at this point
		return
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[31m»\033[0m ",
		HistoryFile:         "/tmp/sUp.tmp",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)

		switch {
		case line == "exit":
			goto exit
		case line == "list":
			printMappingsTable(client)
			break
		case line == "ip":
			printIPTable(client)
			break
		case line == "help":
			usage(l.Stderr())
			break
		case strings.HasPrefix(line, "map"):
			mapPort(line[3:], client)
			break
		default:
			fmt.Println("Command not recognised")
		}
	}
exit:
}

func printMappingsTable(c *sUPnP.IGD) {
	pms, err := c.GetPortMappings()
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Remote", "External Port", "Protocal", "Internal Port", "Internal Client", "Description", "Lease"})

	for _, pm := range pms {
		table.Append(pmToStringMap(pm))
	}

	table.Render()
}

func pmToStringMap(pm *sUPnP.PortMapping) []string {
	return []string{
		pm.RemoteHost,
		strconv.FormatInt(int64(pm.ExternalPort), 10),
		pm.Protocol,
		strconv.FormatInt(int64(pm.InternalPort), 10),
		pm.InternalClient,
		pm.PortMappingDescription,
		strconv.FormatInt(int64(pm.LeaseDuration), 10),
	}
}

func printIPTable(c *sUPnP.IGD) {
	eip, err := c.GetExternalIP()
	if err != nil {
		panic(err)
	}

	iip, err := c.GetInternalIP()
	if err != nil {
		panic(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "IP Address"})

	table.Append([]string{"External", eip})
	table.Append([]string{"Internal", iip})

	table.Render()
}

func mapPort(line string, c *sUPnP.IGD) {
	str := strings.Split(strings.TrimSpace(line), " ")

	if str[0] == "add" {
		// add a port mapping
		str = str[1:]

		remote := strings.Trim(str[0], "\"")
		proto := str[2]
		client := str[4]
		eport, err := strconv.ParseInt(str[1], 10, 16)
		if err != nil {
			panic(err)
		}
		iport, err := strconv.ParseInt(str[3], 10, 16)
		if err != nil {
			panic(err)
		}
		description := strings.Trim(str[5], "\"")
		lease, err := strconv.ParseInt(str[6], 10, 16)
		if err != nil {
			panic(err)
		}

		pm := &sUPnP.PortMapping{
			RemoteHost:             remote,
			ExternalPort:           uint16(eport),
			Protocol:               proto,
			InternalPort:           uint16(iport),
			InternalClient:         client,
			PortMappingDescription: description,
			Enabled:                true,
			LeaseDuration:          uint32(lease),
		}

		c.AddPortMapping(pm)
	} else {
		// remove a port mapping
		str = str[1:]

		remote := strings.Trim(str[0], "\"")
		proto := str[2]
		eport, err := strconv.ParseInt(str[1], 10, 16)
		if err != nil {
			panic(err)
		}

		pm := &sUPnP.PortMapping{
			RemoteHost:   remote,
			ExternalPort: uint16(eport),
			Protocol:     proto,
		}

		c.DeletePortMapping(pm)
	}

	printMappingsTable(c)
}
