package cli

import (
	"errors"
	"fmt"
	"syscall"

	"health_monitor/api"
	"health_monitor/setup"
	"health_monitor/utils"

	"github.com/fatih/color"
)

const (
	helpString = `Usage: command <arguments>
List of commmands:
help			: To view this message
enable <moduleName>	: To enable a module
disable <moduleName>	: To disable a module
status			: To check status of all modules
exit			: To turn off the monitor`
)

func disableModule(argument []string) error {
	if len(argument) > 1 {
		return errors.New("Wrong command, use disable <moduleName>")
	}
	return toggleModule(argument[0], false)
}

func exit(argument []string) error {
	color.Blue("Shutting down monitor gacefully")
	utils.ExitChan <- syscall.SIGINT
	return nil
}

func enableModule(argument []string) error {
	if len(argument) > 1 {
		return errors.New("Wrong command, use" + color.CyanString("enable <moduleName>"))
	}
	return toggleModule(argument[0], true)
}

func help(argument []string) error {
	color.Green("CLI for the OWTF - Health Monitor.")
	color.Cyan(helpString)
	return nil
}

func status(argument []string) error {
	if len(argument) == 0 {
		printHeading("OWTF - Health Monitor Module's Status")
		liveShortStatus()
		diskShortStatus()
		cpuShortStatus()
		ramShortStatus()
		targetShortStatus()
	} else if len(argument) == 1 {
		switch argument[0] {
		case "live":
			liveDetailStatus()
		case "disk":
			diskDetailStatus()
		case "cpu":
			cpuDetailStatus()
		case "ram":
			ramDetailStatus()
		default:
			color.Red("Module not found")
		}
	} //TODO Print if command is wrong
	return nil
}

func liveShortStatus() {
	fmt.Printf("%-35s", "Internet Connectivity (live)")
	moduleWorkingStatus(setup.ModulesStatus.Live, api.LiveStatus().Normal)
}

func liveDetailStatus() {
	printHeading("Status of internet connectivity (live) module")
	fmt.Printf("%-35s", "Internet Connectivity (live) :\t")
	if setup.ModulesStatus.Live {
		moduleStatus := api.LiveStatus()
		if moduleStatus.Normal {
			color.Green("On")
			fmt.Println("You are connected to the internet.")
		} else {
			color.Red("On")
			fmt.Println("You are not connected to the internet.")
		}
	} else {
		color.Cyan("Off")
		fmt.Println("Please turn on the module to monitor internet connectivity status")
	}
}

func diskShortStatus() {
	normal := true
	fmt.Printf("%-35s", "Disk (disk)")
	for _, value := range api.DiskStatus() {
		if value.Status.Inode != 1 || value.Status.Space != 1 {
			normal = false
			break
		}
	}
	moduleWorkingStatus(setup.ModulesStatus.Disk, normal)
}

func diskDetailStatus() {
	printHeading("Status of disk module")
	diskShortStatus()
	if setup.ModulesStatus.Disk {
		printDiskTable()
		fmt.Println("\n")
		printInodeTable()
	}
}

func printDiskTable() {
	fmt.Println("Description of disk blocks available in the system:")
	printLine()
	color.New(color.FgWhite, color.Bold, color.Underline).Printf("| %-30s | %-15s | %-15s |  %%  |\n",
		"Filesystem", "Free Blocks", "Total Blocks")
	if setup.ModulesStatus.Disk {
		for key, value := range api.DiskStatus() {
			colorFunc := color.New(color.FgWhite)
			if value.Status.Space != 1 {
				colorFunc.Add(color.FgRed)
			}
			colorFunc.Printf("| %-30s | %-15d | %-15d | %d%% |\n", key,
				value.Stats.FreeBlocks, value.Const.TotalBlocks,
				percent(value.Stats.FreeBlocks, value.Const.TotalBlocks))
		}
	}
	printLine()
}

func printInodeTable() {
	fmt.Println("Description of disk blocks available in the system:")
	printLine()
	color.New(color.FgWhite, color.Bold, color.Underline).Printf("| %-30s | %-15s | %-15s |  %%  |\n",
		"Filesystem", "Free Inodes", "Total Inodes")
	if setup.ModulesStatus.Disk {
		for key, value := range api.DiskStatus() {
			colorFunc := color.New(color.FgWhite)
			if value.Status.Inode != 1 {
				colorFunc.Add(color.FgRed)
			}
			colorFunc.Printf("| %-30s | %-15d | %-15d | %d%% |\n", key,
				value.Stats.FreeInodes, value.Const.TotalInodes,
				percent(value.Stats.FreeInodes, value.Const.TotalInodes))
		}
	}
	printLine()
}

func printLine() {
	color.New(color.FgWhite, color.Bold, color.Underline).Printf("%76s\n", " ")
}

func percent(value int, total int) int {
	return (value * 100) / total
}

func cpuShortStatus() {
	fmt.Printf("%-35s", "CPU (cpu)")
	moduleWorkingStatus(setup.ModulesStatus.CPU, api.CPUStatus().Status.Normal)
}

func cpuDetailStatus() {
	printHeading("Status of CPU module")
	cpuShortStatus()
	if setup.ModulesStatus.CPU {
		moduleStatus := api.CPUStatus()
		colorFunc := color.New(color.FgWhite)
		if moduleStatus.Status.Normal == false {
			colorFunc.Add(color.FgRed)
		}
		colorFunc.Printf("CPU usage is %f%%\n", moduleStatus.Stats.CPUUsage)
	}
}

func ramShortStatus() {
	fmt.Printf("%-35s", "RAM (ram)")
	moduleWorkingStatus(setup.ModulesStatus.RAM, api.RAMStatus().Status.Normal)
}

func ramDetailStatus() {
	printHeading("Status of RAM module")
	ramShortStatus()
	if setup.ModulesStatus.RAM {
		moduleStatus := api.RAMStatus()
		colorFunc := color.New(color.FgWhite)
		if moduleStatus.Status.Normal == false {
			colorFunc.Add(color.FgRed)
		}
		colorFunc.Printf("RAM usage is %d%%\n", percent(moduleStatus.Stats.FreePhysical,
			moduleStatus.Consts.TotalPhysical))

		colorFunc.Printf("Virtual Memory usage is %d%%\n", percent(moduleStatus.Stats.FreeVirtual,
			moduleStatus.Consts.TotalVirtual))
	}
}

func targetShortStatus() {
	fmt.Printf("%-35s", "OWTF's Targets (target)")
	normal := true

	for _, value := range api.TargetStatus() {
		if value.Scanned {
			if value.Normal == false {
				normal = false
			}
		}
	}

	moduleWorkingStatus(setup.ModulesStatus.Target, normal)
}

func targetDetailStatus() {
	color.Cyan("Detailed status of all the OWTF's target")
	targetShortStatus()
	if setup.ModulesStatus.Target {
		for key, value := range api.TargetStatus() {
			fmt.Printf("%-45s :\t", key)
			if value.Scanned {
				if value.Normal {
					color.Green("Connected")
				} else {
					color.Red("Not connected")
				}
			} else {
				color.Cyan("Not under scan")
			}
		}
	}
}

func printHeading(heading string) {
	color.New(color.Underline, color.Italic, color.FgCyan, color.Bold).Println(heading)
}

func moduleWorkingStatus(status bool, workingStatus bool) {
	fmt.Print(":\t")
	if status {
		if workingStatus {
			color.Green("On")
		} else {
			color.Red("On")
		}
	} else {
		color.Cyan("Off")
	}
}

func toggleModule(module string, state bool) error {
	if doesModuleExists(module) {
		utils.SendModuleStatus(module, state)
		return nil
	} else {
		return errors.New("Specified module not found, allowed modules " + color.New(color.FgCyan).SprintFunc()(utils.Modules))
	}
}

func doesModuleExists(module string) bool {
	for _, workingModule := range utils.Modules {
		if workingModule == module {
			return true
		}
	}
	return false
}