package main

import (
    "strings"
    "os/exec"
    "runtime"
    "os"
)

func command_help() (string) {
    var raw_output = []string{"",
        "Available internal commands:",
        "    '!help' - Displays this help output",
        "    '!clear' - Clears the screen",
        "    '!exit' - Terminates the session and closes the RCON client",
        "",
        "IW4x RCON help (not prefixed with !):",
        "    'cmdlist' - List every IW4x server executable command",
        "    'status' - Basic player information",
        "",
        "IW4x common commands (not prefixed with !):",
        "    'map <map>' - Change the map to <map> (Example: 'map mp_rust')",
        "    'set g_gametype <gamemode>' - Change current gamemode to <gamemode>",
        "    'clientkick <ID>' - Kick player with an ID of <ID> (acquire from 'status')",
        "    'say <message>' - Send message to game chat as the server",
        "    'banClient <ID>' - Ban player with an ID of <ID>",
        "    'banUser <Username>' - Ban player with a username of <Username>",
        "    'tempBanClient <ID>' - Temporarily ban player with an ID of <ID>",
        "    `tempBanUser <Username>` - Temporarily ban player with a username of <Username>",
        "    'unbanUser <Username>' - Unban player with a username of <Username>",
        "    'unbanClient <ID>' - Unban player with an ID of <ID>",
        "    'muteClient <ID>' - Mute player with an ID of <ID>",
        "    'unmute <ID>' - Unmute player with an ID of <ID>",
        "    ...see cmdlist...",
        "",
        "When running an IW4x command with no arguments, in general, usage information will be passed back.",
        "\n"}

    output := strings.Join(raw_output[:], "\n")
    
    return output
}

func command_clear() {
    var cmd *exec.Cmd
    if runtime.GOOS == "windows" {
        cmd = exec.Command("cls")
    } else {
        cmd = exec.Command("clear")
    }
    cmd.Stdout = os.Stdout
    cmd.Run()
}
