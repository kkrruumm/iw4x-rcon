package main

import (
    "bufio"
    "time"
    "net"
    "fmt"
    "strings"
    "os"
    "flag"
    "crypto"
    "crypto/rsa"
    "crypto/sha512"
    "crypto/rand"
    "crypto/x509"
    "encoding/pem"
    "encoding/binary"
)

type iw4x_client struct {
    Connection net.Conn
    PrivateKey *rsa.PrivateKey
    Password string
    SafeMode bool
}

func spawn_session(addr string, password string, key_path string) (*iw4x_client, error) {
    connection, err := net.Dial("udp", addr)
    if err != nil {
        return nil, err
    }

    c := &iw4x_client{}

    if key_path != "" {
        key_bytes, err := os.ReadFile(key_path)
        if err != nil {
            return nil, fmt.Errorf("failed to read keyfile: %w", err)
        }

        block, _ := pem.Decode(key_bytes)
        if block == nil {
            return nil, fmt.Errorf("failed to parse PEM block: %w", err)
        }

        private_key, err := x509.ParsePKCS8PrivateKey(block.Bytes) // openssl default as of the time of writing
        if err != nil {
            return nil, fmt.Errorf("failed to parse RSA key: %w", err)
        }

        var rsa_private_key *rsa.PrivateKey
        rsa_private_key = private_key.(*rsa.PrivateKey)
        
        c.PrivateKey = rsa_private_key
        c.SafeMode = true
    } else {
        c.Password = password
        c.SafeMode = false
    }

    c.Connection = connection
    return c, nil
}

func (c *iw4x_client) send_command(command string) (string, error) {
    var packet []byte
    if c.SafeMode {
        hashed_command := sha512.Sum512([]byte(command)) // iw4x expects sha512 for command hashing
        signature, err := rsa.SignPKCS1v15(rand.Reader, c.PrivateKey, crypto.SHA512, hashed_command[:]) // iw4x expects PKCS1v15 padding
        if err != nil {
            return "", fmt.Errorf("failed to sign command: %w", err)
        }

        // the expected packet structure from this: \xff\xff\xff\xffrconSafe <command_tag><command_length><command> <signature_tag><signature_length><signature>
        // yes this is protobuf :| (todo: consider protobuf lib)
        packet = append(packet, []byte("\xff\xff\xff\xffrconSafe ")...) // header, quake hours

        packet = append(packet, 0x0A) // add command tag
        b := make([]byte, binary.MaxVarintLen64) // create a buffer to store a slice of bytes to store variable-length ints

        // calculate length of command, convert it to a uint64, write it into buffer "b"
        // b[:...] is used here because we might only need XYZ amount of bytes- so we clip the buffer to not add junk to the packet
        packet = append(packet, b[:binary.PutUvarint(b, uint64(len(command)))]...)
        packet = append(packet, []byte(command)...) // add command itself

        packet = append(packet, 0x12) // signature tag
        b2 := make([]byte, binary.MaxVarintLen64)
        packet = append(packet, b2[:binary.PutUvarint(b2, uint64(len(signature)))]...)
        packet = append(packet, []byte(signature)...) // the actual signature
    } else { // unsafe rcon, simple packet structure
        packet = []byte(fmt.Sprintf("\xff\xff\xff\xffrcon %s %s", c.Password, command))
    }

    // send the command
    _, err := c.Connection.Write([]byte(packet))
    if err != nil {
        return "", fmt.Errorf("failed to send command to server: %w", err)
    }

    // this is our timeout for the command
    // if the server takes longer than this to reply, need to get rid of it
    c.Connection.SetReadDeadline(time.Now().Add(10 * time.Second))
    buf := make([]byte, 12288) // pre-allocated chunk of memory to store the response, 12KB

    // gos `net` stack will take care of fragmented packets here so this doesn't need to loop
    raw_response, err := c.Connection.Read(buf) // read response into that chunk
    if err != nil {
        return "", fmt.Errorf("failed to read server response: %w", err)
    }

    // cleaning up the response to be output
    response := string(buf[:raw_response])
    // clean up engine header stuff being returned
    response = strings.TrimPrefix(response, "\xff\xff\xff\xffprint\n")
    
    // make sure response actually has a length before returning success
    if len(response) == 0 {
        return "", fmt.Errorf("%w", err)
    }

    return strings.TrimSpace(response), nil
}

func (c *iw4x_client) Close() error {
    return c.Connection.Close()
}

func main() {
    // variables for command line argument assignment
    ip_address := flag.String("i", "", "IP Address")
    port := flag.String("p", "28960", "Server Port")
    key_path := flag.String("k", "", "Path to RSA private key for Safe-RCON")
    rcon_pass := flag.String("pass", "", "Insecure RCON Password")
    flag.Parse()

    if *ip_address == "" || (*key_path != "" && *rcon_pass != "") {
        fmt.Print("Usage requirements:\n")
        fmt.Print("  Safe-RCON: -i, -k\n")
        fmt.Print("  Insecure:  -i, -pass\n\n")
        flag.Usage()
        return
    }

    session, err := spawn_session(*ip_address+":"+*port, *rcon_pass, *key_path)
    if err != nil {
        fmt.Printf("failed to spawn session with server: %s\n", err)
        return
    }
    defer session.Close() // do not close the connection until main returns, this lets us avoid spamming connections

    fmt.Print("Connection started.\n")
    fmt.Print("You may enter '!help' for information on using this command line.\n")
    reader := bufio.NewReader(os.Stdin)

    // input handling
    for {
        if session.SafeMode {
            fmt.Printf("(%s:%s) Safe-RCON> ", *ip_address, *port)
        } else {
            fmt.Printf("(%s:%s) Insecure-RCON> ", *ip_address, *port)
        }

        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        switch internal_command := input; internal_command {
        case "!help":
            fmt.Printf(command_help())
            continue
        case "!clear":
            command_clear()
            continue
        case "!exit":
            fmt.Print("Goodbye!\n")
            return
        case "":
            continue // if the user just hits enter we shouldnt send that to the server
        }

        response, err := session.send_command(input)
        if err != nil {
            fmt.Printf("error running command: %s\n", err)
        }

        fmt.Printf("\n%s\n\n", response)
        time.Sleep(200 * time.Millisecond) // attempt to prevent command timeouts from spam
    }
}
