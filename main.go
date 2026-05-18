package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var trayIcon []byte

func main() {
	if hasLaunchArg("--core") {
		if err := runCoreMain(); err != nil {
			log.Fatal(err)
		}
		return
	}

	recoverBrokenSingleInstance("com.novaproxy.desktop")

	app := NewApp()

	wailsApp := application.New(application.Options{
		Name:        "novaproxy",
		Description: "NovaProxy - Cloudflare IP Shaper",
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(assets),
		},
		Services: []application.Service{
			application.NewService(app),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.novaproxy.desktop",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				app.RevealMainWindow()
			},
			ExitCode: 0,
		},
	})

	app.wailsApp = wailsApp

	tray := wailsApp.SystemTray.New()
	tray.SetIcon(trayIcon)
	tray.SetDarkModeIcon(trayIcon)
	tray.SetTooltip("NovaProxy")
	app.systemTray = tray

	trayMenu := application.NewMenu()
	trayMenu.Add("Dashboard").OnClick(func(ctx *application.Context) {
		app.RevealMainWindow()
	})
	trayMenu.AddSeparator()

	proxyLabel := "Proxy: Off"
	if app.IsProxyRunning() {
		proxyLabel = "Proxy: On"
	}
	app.proxyItemV3 = trayMenu.AddCheckbox(proxyLabel, app.IsProxyRunning())
	app.proxyItemV3.OnClick(func(ctx *application.Context) {
		app.runSafeAsync("tray proxy toggle", func() {
			if app.IsProxyRunning() {
				_ = app.StopProxy()
			} else {
				_ = app.StartProxy()
			}
		})
	})

	systemProxyLabel := "System Proxy: Off"
	if app.GetSystemProxyStatus().Enabled {
		systemProxyLabel = "System Proxy: On"
	}
	app.systemProxyItemV3 = trayMenu.Add(systemProxyLabel)
	app.systemProxyItemV3.OnClick(func(ctx *application.Context) {
		app.runSafeAsync("tray system proxy toggle", func() {
			if app.GetSystemProxyStatus().Enabled {
				_ = app.DisableSystemProxy()
				return
			}
			if !app.IsProxyRunning() {
				if err := app.StartProxy(); err != nil {
					return
				}
			}
			_ = app.EnableSystemProxy()
		})
	})

	trayMenu.AddSeparator()
	trayMenu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.QuitApp()
	})

	tray.SetMenu(trayMenu)
	app.trayMenuV3 = trayMenu

	app.mainWindow = wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "main",
		Title:            "NovaProxy",
		Width:            1024,
		Height:           768,
		URL:              "/",
		Frameless:        true,
		Hidden:           app.ShouldStartHidden(),
		BackgroundColour: application.NewRGB(27, 38, 54),
	})
	tray.OnClick(func() {
		app.RevealMainWindow()
	})

	err := wailsApp.Run()
	if err != nil {
		log.Fatal(err)
	}
}
