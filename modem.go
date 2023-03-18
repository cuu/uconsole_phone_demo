package main

import (
	//"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	//"sync"
	"errors"
	"github.com/maltegrosse/go-modemmanager"
	"github.com/visualfc/atk/tk"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"strings"
	"time"
)

var number string
var current_call = make(chan modemmanager.Call, 1)
var global_modem modemmanager.Modem

func play_ring(done chan bool) {
	fmt.Println("play_ring now")
	f, err := os.Open("telephone-ring-03a.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false}

	speaker.Play(ctrl)

	for {
		select {
		case <-done:
			return
		}
		/*
			speaker.Lock()
			ctrl.Paused = !ctrl.Paused
			speaker.Unlock()
		*/
	}
}

func pipeNewCall(call modemmanager.Call) {
	select {
	case current_call <- call: // set current call
		fmt.Println("Set current call to channel")
	default:
		fmt.Println("set call channel failed!")
	}

}

func GetCallNumber(call modemmanager.Call) string {
	str, err := call.GetNumber()
	if err != nil {
		return ""
	}
	return str
}

func InitModem(window *Window) modemmanager.Modem {

	mmgr, err := modemmanager.NewModemManager()
	if err != nil {
		log.Fatal(err.Error())
	}
	version, err := mmgr.GetVersion()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("ModemManager Version: ", version)
	modems, err := mmgr.GetModems()
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, modem := range modems {
		fmt.Println("ObjectPath: ", modem.GetObjectPath())
		go listenToModemVoiceCallAdded(modem, window)
		go listenToModemMessagingAdded(modem, window)
		global_modem = modem

		break
	}
	maxVolume()
	syncSms(window)
	return global_modem

}

func createCall(window *Window, number string) {

	//	time.Sleep(time.Second * time.Duration(3))
	//	InitModem(window)

	voice, err := global_modem.GetVoice()
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(voice.GetObjectPath())

	if call, err := voice.CreateCall(number); err == nil {
		call.Start()
		SetDecline(window,true)
		SetCall(window,false)
	} else {
		log.Fatal("create call failed", err)
	}
}

func rejectCall() error {

	select {
	case call := <-current_call:
		//fmt.Println("get call from channel")
		state,err := call.GetState()
		if err != nil {
			log.Println(err)
			return nil
		}
		if state != modemmanager.MmCallStateActive {
			voice, _:= global_modem.GetVoice()
			voice.HangupAll()
		}else{
			err := call.Hangup() //requires sudo
			if err != nil {
				fmt.Println(err)
			}
		}
	default:
		fmt.Println("no call object")
	}

	return nil
}

func acceptCall() error {

	select {
	case call := <-current_call:
		//fmt.Println("get call from channel")
		err := call.Accept() //requires sudo
		if err != nil {
			fmt.Println(err)
		}
	default:
		fmt.Println("no call object")
	}

	return nil
}

func getSms(idx int) modemmanager.Sms {
        msg,err := global_modem.GetMessaging()
        if err != nil {
                log.Fatal(err.Error())
        }

        fmt.Println(msg.GetObjectPath())

        allsms,err := msg.List()
	if err == nil {
		for i,v := range allsms {
			if idx == i {
				return v
			}
		}
	}

	return nil
}

func syncSms(window *Window) {
	msg,err := global_modem.GetMessaging()
	if err != nil {
		log.Fatal(err.Error())
	}
	
	fmt.Println(msg.GetObjectPath())
	ClearLogs(window)
	allsms,err := msg.List()
	fmt.Println("allsms :",len(allsms))

	if err == nil {
		for i,v := range allsms {
			state ,_ := v.GetState()
			if state.String() == "Receiving" || state.String() == "Sending" {
				item_text := fmt.Sprintf("#%d# %s",i+1,state.String())
				AddLog(window,item_text)
			}else {
				txt, err2 := v.GetText()
				if err2 == nil {
					item_text := fmt.Sprintf("#%d# %s",i+1,txt)
					AddLog(window,item_text)
				}
			}
			//fmt.Println(txt,state.String())
		}
	}

}

func sendMessage(window *Window) error {
	number := window.NumberEntry.Text()
	msgText := window.MsgEntry.Text()
	sendBtn := window.SendBtn	
	if len(number) > 2 && len(msgText) > 0 {
		msg, err := global_modem.GetMessaging()
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(msg.GetObjectPath())

		if sms, err := msg.CreateSms(number, msgText); err == nil {
			sms.Send()
		} else {
			log.Fatal("create sms failed", err)
		}
	
	}else{
		return errors.New("no msg")
	}
	
	sendBtn.SetState(tk.StateNormal)
	return nil
}

func maxVolume() error {
	atcmd := "AT+CLVL=5"
	_, err := global_modem.Command(atcmd, 1)
	return err
}
func listenToModemMessagingAdded(modem modemmanager.Modem, window *Window) {
	msg, err := modem.GetMessaging()
	if err != nil {
		log.Fatal(err.Error())
	}
	c := msg.SubscribeAdded()
	fmt.Println("start listening sms....")
	for v := range c {
		fmt.Println("SmsAdded ", v)
		fmt.Println(reflect.TypeOf(v))
		fmt.Println("name", v.Name)
		fmt.Println("path", v.Path)
		fmt.Println("body", v.Body)
		fmt.Println("listenToModemMessagingAdded sender", v.Sender)

		if strings.Contains(v.Name, modemmanager.ModemMessagingSignalAdded) == true {
			syncSms(window)
			go func() {
				cur_sms,_,_ := msg.ParseAdded(v)
				if cur_sms == nil {
					return
				}
				state_changed := cur_sms.SubscribePropertiesChanged()
				for val := range state_changed {
					cur_sms.ParsePropertiesChanged(val)
					state,_ := cur_sms.GetState()
					if state.String() == "Received" || state.String() == "Sent" || state.String() == "Unknown" {
						break
					}
				}
				cur_sms.Unsubscribe()
				syncSms(window)
			}()
		}
	}

}
func listenToModemVoiceCallAdded(modem modemmanager.Modem, window *Window) {
	// listen new calls
	voice, err := modem.GetVoice()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(voice.GetObjectPath())
	c := voice.SubscribeCallAdded()
	fmt.Println("start listening ....")

	for v := range c {
		fmt.Println("CallAdded ", v)
		fmt.Println(reflect.TypeOf(v))
		fmt.Println("name", v.Name)
		fmt.Println("path", v.Path)
		fmt.Println("body", v.Body)
		fmt.Println("listenToModemVoiceCallAdded sender", v.Sender)

		if strings.Contains(v.Name, modemmanager.ModemVoiceSignalCallAdded) == true {
			SetStatusBar(window,"uConsole")
			calls, err := voice.ParseCallAdded(v)
			if err == nil {
				fmt.Println("newCall()")

				state, err := calls.GetState()
				if err == nil {
					fmt.Println("newCall()", state)
					if state == modemmanager.MmCallStateRingingIn {
						pipeNewCall(calls)
						SetAccept(window,true)
						SetDecline(window,true)

						number = GetCallNumber(calls)

						ch_stop_ring := make(chan bool)
						go play_ring(ch_stop_ring)
						state_changed := calls.SubscribeStateChanged()
						fmt.Println("newCall() wait call state change")
						ch_count := make(chan bool)
						for val := range state_changed {
							fmt.Println(" call.SubscribeStateChanged ", val)
							fmt.Println(reflect.TypeOf(val))
							fmt.Println("name", val.Name)
							fmt.Println("path", val.Path)
							fmt.Println("body", val.Body)
							fmt.Println("sender", val.Sender)
							oldState, newState, reason, err := calls.ParseStateChanged(val)
							if err == nil {
								log.Println("oldState:", oldState)
								log.Println("newState:", newState)
								log.Println("reason:", reason)

								if newState == modemmanager.MmCallStateActive {
									SetStatusBar(window, "Talking....")
									StartCount(window, ch_count)
								}
								if newState == modemmanager.MmCallStateTerminated {
									ch_stop_ring <- true
									ch_count <- true
									SetStatusBar(window, fmt.Sprintf("%s Call end", number))
									break
								}
							}
						}
						calls.Unsubscribe()
						ResetAcceptAndDecline(window)
					}
					if state == modemmanager.MmCallStateUnknown || state == modemmanager.MmCallStateRingingOut {
						pipeNewCall(calls)
						SetAccept(window,false)
						SetDecline(window,true)
						//change label
						SetStatusBar(window, "Ringing...")
						state_changed := calls.SubscribeStateChanged()
						fmt.Println("newCallOut() wait call state change")
						ch_count := make(chan bool,1)
						for val := range state_changed {
							fmt.Println(" call.SubscribeStateChanged ", val)
							fmt.Println(reflect.TypeOf(val))
							fmt.Println("name", val.Name)
							fmt.Println("path", val.Path)
							fmt.Println("body", val.Body)
							fmt.Println("sender", val.Sender)
							oldState, newState, reason, err := calls.ParseStateChanged(val)
							if err == nil {
								log.Println("oldState:", oldState)
								log.Println("newState:", newState)
								log.Println("reason:", reason)

								if newState == modemmanager.MmCallStateActive {
									fmt.Println("now talk")
									SetStatusBar(window, "Talking....")
									go StartCount(window, ch_count)
									SetDecline(window,true)
									SetAccept(window,false)
									
								}
								if newState == modemmanager.MmCallStateWaiting {
									SetStatusBar(window, "Waiting to pick...")
								}

								if newState == modemmanager.MmCallStateTerminated {
									SetStatusBar(window, "Call rejected...")
									SetCall(window,true)	
									log.Println(fmt.Sprintf("Call %s end", GetCallNumber(calls)))
									ch_count <- true
									break
								}
							}
						}
						calls.Unsubscribe()
						ResetAcceptAndDecline(window)
					}
				} else {
					fmt.Println(err)
				}
			} else {
				fmt.Println(err)
			}
		}

	}
}
