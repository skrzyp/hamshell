package main

import (
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/pterm/pterm"
	"github.com/skrzyp/hamshell/qrz"
	"github.com/spf13/viper"
)

type Command struct {
	Prefix      string
	Description string
	Usage       string
	ArgsMin     int
	ArgsMax     int
	Method      func([]string)
}

var Commands []Command

func qrzCommand(args []string) {
    	loginQRZ := viper.GetString("qrz.login")
    	if loginQRZ == "" {
        	pterm.Error.Println("No QRZ login set, edit qrz.login in hamshell config.")
        	return
    	}
    	passwordQRZ := viper.GetString("qrz.password")
    	if passwordQRZ== "" {
        	pterm.Error.Println("No QRZ password set, edit qrz.password in hamshell config.")
        	return
    	}
    	s, err := qrz.New(loginQRZ, passwordQRZ)
	if err != nil {
		pterm.Warning.Printfln("Error logging to qrz.com: %s", err)
		return
	}
	c, err := s.GetCall(args[0])
	if err != nil {
		pterm.Warning.Printfln("Error getting data for callsign %s: %s", args[0], err)
		return
	}
	c.Print()
}

func helpCommand(args []string) {
	if len(args) < 1 {
		pterm.Println("Available commands: ")
		for _, cmd := range Commands {
			pterm.Printfln("%s %s", cmd.Prefix, cmd.Description)
		}
		return
	} else {
		for _, cmd := range Commands {
			if args[0] == cmd.Prefix {
				pterm.Println(cmd.Description)
				pterm.Println("Usage: " + cmd.Usage)
				return
			}
		}
		pterm.Warning.Printfln("Unknown command: %s", args[0])
	}
}

func completor(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func executor(input string) {
	if input != "" {
		args := strings.Fields(input)
		var c Command
		for _, cmd := range Commands {
			if args[0] == cmd.Prefix {
				c = cmd
				break
			}
		}
		if c.Description != "" {
			if len(args) < c.ArgsMin+1 || (len(args) > c.ArgsMax+1) {
				pterm.Warning.Printfln("Usage: %s", c.Usage)
			} else {
				c.Method(args[1:])
			}
		} else {
			pterm.Error.Printfln("Unknown command %s", args[0])
		}
	}
}

func main() {
	Commands = []Command{
		{"qrz", "Get info about specific callsign", "qrz [callsign]", 1, 1, qrzCommand},
		{"help", "Get help about specific command", "help [command]", 0, 1, helpCommand},
	}

	pterm.Info.Println("Welcome to hamshell.")

	viper.SetConfigName("hamshell")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("callsign", "")
	viper.SetDefault("qrz.login", "")
	viper.SetDefault("qrz.password", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
    			pterm.Warning.Println("No config file found, writing default config")
    			viper.SafeWriteConfig()
		} else {
			pterm.Error.Printfln("Error reading config file: %s", err)
			return
		}
	}

	p := prompt.New(
		executor,
		completor,
		prompt.OptionPrefix("> "),
	)
	p.Run()
}
