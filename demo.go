package main

import (
	"fmt"
	"time"
	"log"
	"github.com/visualfc/atk/tk"
)
type SmsWindow struct {
	*tk.Window
	Msg *tk.TextEx
}

type Window struct {
	*tk.Window
	smsWin *SmsWindow
	NumberEntry *tk.Entry
	MsgEntry    *tk.Entry
	CallBtn     *tk.Button
	SendBtn     *tk.Button
	AcceptBtn   *tk.Button
	DeclineBtn  *tk.Button
	Logs        *tk.ListBoxEx
	TimeLabel   *tk.Label
	StatusLabel	*tk.Label
}

func secondsToMinutes(inSeconds int) string {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	str := fmt.Sprintf("%02d:%02d", minutes, seconds)
	return str
}

func StartCount(window *Window, done chan bool) {
	cnt := 0

	for {
		select {
		case <-done:
			tk.Async(func() {
				window.TimeLabel.SetText("00:00")
			})
			return
		case <-time.After(1 * time.Second):
			cnt += 1
			tk.Async(func() {
				lbl_str := secondsToMinutes(cnt)
				fmt.Println(lbl_str)
				window.TimeLabel.SetText(lbl_str)
			})
		}
	}
}

func ClearLogs(window *Window) {
	tk.Async(func() {
		cnt := window.Logs.ItemCount()
		if cnt > 0 {
			window.Logs.RemoveItemRange(0,cnt)
		}
	})
}
func AddLog(window *Window, logs string) {
	tk.Async(func() {
		cnt := window.Logs.ItemCount()
		window.Logs.InsertItem(cnt, logs)
	})
}
func SetStatusBar(window *Window, txt string) {
	tk.Async(func() {
		window.StatusLabel.SetText(txt)
	})
}

func SetAccept(window *Window, state bool) {
	if state == true {
		tk.Async(func() {
			window.AcceptBtn.SetState(tk.StateNormal)
		})
	}else {
		tk.Async(func() {
			window.AcceptBtn.SetState(tk.StateDisable)
		})
	}
}

func SetDecline(window *Window,state bool) {
    if state == true {
		tk.Async(func() {
	        window.DeclineBtn.SetState(tk.StateNormal)
		})
    }else {
		tk.Async(func() {
	        window.DeclineBtn.SetState(tk.StateDisable)
		})
    }
}
func SetCall(window *Window, state bool) {
	if state == true {
		tk.Async(func() {
			window.CallBtn.SetState(tk.StateNormal)
		})
	} else {
		tk.Async(func() {
			window.CallBtn.SetState(tk.StateDisable)
		})
	}	
}

func ResetAcceptAndDecline(window *Window) {
	tk.Async(func() {
		window.AcceptBtn.SetState(tk.StateDisable)
		window.DeclineBtn.SetState(tk.StateDisable)
	})
}

func NewSmsWindow() *SmsWindow {
	smsWin := &SmsWindow{}

	smsWin.Window = tk.NewWindow()
	smsWin.Msg = tk.NewTextEx(smsWin.Window)
	

	smsWin.Window.OnClose(func() bool {
		smsWin.Hide()
		return false
	})
	return smsWin
}

func NewWindow() *Window {
	mw := &Window{}
	mw.Window = tk.RootWindow()

	mbar := tk.NewMenu(mw)
	mw.SetMenu(mbar)

	file := mbar.AddNewSubMenu("File")
	file.AddAction(tk.NewActionEx("Open", func() {
		fmt.Println("Open")
	}))
	file.AddSeparator()
	file.AddAction(tk.NewActionEx("Quit", func() {
		tk.Quit()
	}))

	lbl := tk.NewLabel(mw, "Enter Phone Number:")

	tk.Place(lbl, tk.PlaceAttrX(10), tk.PlaceAttrY(5))

	time_lbl := tk.NewLabel(mw, "00:00")
	time_lbl.SetFont(tk.NewUserFont("Fixed", 14))
	tk.Place(time_lbl, tk.PlaceAttrX(340), tk.PlaceAttrY(27))

	number_entry := tk.NewEntry(mw)
	number_entry.SetWidth(26)
	number_entry.SetFont(tk.NewUserFont("Fixed", 14))

	tk.Place(number_entry, tk.PlaceAttrX(10), tk.PlaceAttrY(25))

	lbl2 := tk.NewLabel(mw, "Messages")
	tk.Place(lbl2, tk.PlaceAttrX(10), tk.PlaceAttrY(60))

	msg_entry := tk.NewEntry(mw)
	msg_entry.SetWidth(41)
	tk.Place(msg_entry, tk.PlaceAttrX(10), tk.PlaceAttrY(80))

	logs := tk.NewListBoxEx(mw)
	logs.SetWidth(46)
	logs.SetHeight(8)
	logs.ShowXScrollBar(false)
	tk.Place(logs, tk.PlaceAttrX(10), tk.PlaceAttrY(110))

	logs.OnSelectionChanged( func() {
		curs := logs.SelectedIndexs()
		if len(curs) > 0 {
			sms := getSms(curs[0])
			sms_txt,_ := sms.GetText()
			sms_from_number, _ := sms.GetNumber()
			sms_state,_ := sms.GetState()
			timestamp,_ := sms.GetTimestamp()
			timestamp_txt := timestamp.Format("2006-01-02 15:04:05")	
			formatted_txt := fmt.Sprintf("%s %s %s:%s",sms_state,sms_from_number,timestamp_txt,sms_txt)
			fmt.Println(formatted_txt)
			mw.smsWin.SetTitle("SMS")
			mw.smsWin.Msg.SetText(formatted_txt)
			mw.smsWin.Center(mw)
			mw.smsWin.ShowNormal()
		}	
	})

	mw.Logs = logs

	b1 := tk.NewButton(mw, "Call")
	b1.SetWidth(7)
	b1.SetUnder(0)
	b1.OnCommand(func() {
		number := number_entry.Text()
		if len(number) > 2 {
			fmt.Println("try to call:", number)
			createCall(mw,number)
		}

	})
	b2 := tk.NewButton(mw, "Send")
	b2.SetWidth(6)
	b2.SetUnder(0)
	b2.OnCommand(func() {
    	number := number_entry.Text()
        if len(number) > 2 {
        	fmt.Println("try to send sms:", number)
			msg_text := msg_entry.Text()
			if len(msg_text) > 0 {
				fmt.Println(fmt.Sprintf("send %s to %s",msg_text,number))
				b2.SetState(tk.StateDisable)
				sendMessage(mw)
			}
				
		}	
	})
	b3 := tk.NewButton(mw, "Accept")
	b3.SetWidth(6)
	b3.SetUnder(0)
	b3.OnCommand(func() {
		acceptCall()
	})

	b4 := tk.NewButton(mw, "Decline")
	b4.SetWidth(6)
	b4.SetUnder(0)
	b4.OnCommand(func() {
		rejectCall()
	})
	b3.SetState(tk.StateDisable)
	b4.SetState(tk.StateDisable)

	//b1.SetText("End Call")

	tk.Place(b1, tk.PlaceAttrX(230), tk.PlaceAttrY(260))
	tk.Place(b2, tk.PlaceAttrX(317), tk.PlaceAttrY(260))
	tk.Place(b3, tk.PlaceAttrX(138), tk.PlaceAttrY(260))
	tk.Place(b4, tk.PlaceAttrX(50), tk.PlaceAttrY(260))
	
	status_lbl := tk.NewLabel(mw, "uConsole")
	tk.Place(status_lbl, tk.PlaceAttrX(0), tk.PlaceAttrY(290))

	mw.MsgEntry = msg_entry
	mw.NumberEntry = number_entry
	mw.TimeLabel = time_lbl
	mw.StatusLabel = status_lbl

	mw.CallBtn = b1
	mw.SendBtn = b2
	mw.AcceptBtn = b3
	mw.DeclineBtn = b4

	mw.ResizeN(410, 310)

	//AddLog(mw, "Test")
	mw.smsWin = NewSmsWindow()
	mw.smsWin.ResizeN(410,150)
	mw.smsWin.SetResizable(false, false)
	mw.smsWin.SetTopmost(true)
	mw.smsWin.Hide()	
	mw.smsWin.Center(mw)
	return mw
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)

	tk.MainLoop(func() {
		mw := NewWindow()
		mw.SetTitle("uConsole Phone Demo")
		mw.Center(nil)
		mw.SetResizable(false, false)
		mw.ShowNormal()

		InitModem(mw)
	})
}
