package ram

import (
	"encoding/json"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"health_monitor/setup"
	"health_monitor/utils"
)

type (
	// Status holds the status of the RAM after the scan
	Status struct {
		Normal bool
	}
	//Info holds the information of RAM's status, contants and
	// stats after the scan
	Info struct {
		Status Status
		Stats  MemoryStat
		Consts MemoryConst
	}
)

var (
	ramInfo Info
	logFile *os.File
	conf    *Config
)

//RAM is the driver function of this module for monitor
func RAM(status <-chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	var logFileName = path.Join(setup.ConfigVars.HomeDir, "ram.log")

	logFile = utils.OpenLogFile(logFileName)
	defer logFile.Close()

	utils.ModuleLogs(logFile, "Running with "+conf.Profile+" profile")
	ramInfo.Consts.InitMemoryConst()
	checkRAM()
	for {
		select {
		case <-status:
			utils.ModuleLogs(logFile, "Recieved signal to turn off. Signing off")
			return
		case <-time.After(time.Millisecond * time.Duration(conf.RecheckThreshold)):
			checkRAM()
			runtime.Gosched()
		}
	}
}

func checkRAM() {
	ramInfo.Stats.LoadMemoryStats()
	if ramInfo.Stats.FreePhysical < conf.RAMWarningLimit {
		ramInfo.Status.Normal = false
		utils.ModuleLogs(logFile, "Ram is being used over the warning limit")
	} else {
		ramInfo.Status.Normal = true
		utils.ModuleLogs(logFile, "Ram usage is normal")
	}
}

// GetStatus function is getter funtion for the ramInfo to send status
// of ram monitor
func GetStatus() Info {
	return ramInfo
}

//GetConfJSON returns the json byte array of the module's config
func GetConfJSON() []byte {
	data, err := json.Marshal(LoadConfig())
	if err != nil {
		utils.ModuleError(logFile, err.Error(), "[!] Check the conf struct")
	}
	return data
}

// GetStatusJSON function retuns the json string of the ramInfo struct
func GetStatusJSON() []byte {
	data, err := json.Marshal(ramInfo)
	if err != nil {
		utils.ModuleError(logFile, err.Error(), "[!] Check the ramInfo struct")
	}
	return data
}

//Init is the initialization function of the module
func Init() {
	conf = LoadConfig()
}