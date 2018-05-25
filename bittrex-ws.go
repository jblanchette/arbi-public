package main

import (
    "encoding/json"
    "fmt"
    "time"
    "errors"
    "github.com/avdva/go-bittrex"
    "github.com/thebotguys/signalr"
)

const (
    WS_BASE     = "socket.bittrex.com" // Bittrex WS API endpoint
    WS_HUB      = "CoreHub"            // SignalR main hub
)

func doAsyncTimeout(f func() error, tmFunc func(error), timeout time.Duration) error {
    errs := make(chan error)
    go func() {
        err := f()
        select {
        case errs <- err:
        default:
            if tmFunc != nil {
                tmFunc(err)
            }
        }
    }()
    select {
    case err := <-errs:
        return err
    case <-time.After(timeout):
        return errors.New("operation timeout")
    }
}

func parseDeltas(messages []json.RawMessage, dataCh chan<- bittrex.SummaryState, markets ...string) error {
    all := make(map[string]struct{})
    for _, market := range markets {
        all[market] = struct{}{}
    }

    for _, msg := range messages {
        var d bittrex.ExchangeDelta
        if err := json.Unmarshal(msg, &d); err != nil {
            return err
        }

        for _, v := range d.Deltas {
            if _, ok := all[v.MarketName]; ok {
                dataCh <- v
            }
        }
    }

    return nil
}


func (e BittrexExchange) JSubscribeSummaryUpdate(dataCh chan<- bittrex.SummaryState, stop <-chan bool, markets ...string) error {
    const timeout = 5 * time.Second
    client := signalr.NewWebsocketClient()
    client.OnClientMethod = func(hub string, method string, messages []json.RawMessage) {
        if hub != WS_HUB || method != "updateSummaryState" {
            return
        }

        parseDeltas(messages, dataCh, markets...)
    }

    connect := func() error { return client.Connect("https", WS_BASE, []string{WS_HUB}) }
    handleErr := func(err error) {
        if err == nil {
            client.Close()
        } else {
            //should maybe panic or something here?
            fmt.Println(err.Error())
        }
    }

    if err := doAsyncTimeout(connect, handleErr, timeout); err != nil {
        return err
    }
    defer client.Close()

    fmt.Println("Subscribing to summary deltas")
    _, err := client.CallHub(WS_HUB, "SubscribeToSummaryDeltas")
    if err != nil {
        fmt.Println("got err")
        fmt.Println(err)
    }

    select {
    case <-stop:
    case <-client.DisconnectedChannel:
    }

    return nil
}
