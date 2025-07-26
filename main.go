package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type ExchangeResponse struct {
	Result           string  `json:"result"`
	Documentation    string  `json:"documentation"`
	TermsOfUse       string  `json:"terms_of_use"`
	TimeLastUpdate   string  `json:"time_last_update_utc"`
	TimeNextUpdate   string  `json:"time_next_update_utc"`
	BaseCode         string  `json:"base_code"`
	TargetCode       string  `json:"target_code"`
	ConversionRate   float64 `json:"conversion_rate"`
	ConversionResult float64 `json:"conversion_result"`
}

type model struct {
	step        int
	amountStr   string
	amount      float64
	fromIndex   int
	toIndex     int
	selectingTo bool
	showResult  bool
	result      float64
	err         error
}

var currencies = []string{"EUR", "RON", "GBP", "AED", "RUB"}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func convert(amount float64, from, to string) (float64, error) {
	url := "https://v6.exchangerate-api.com/v6/" + "your key" +"/pair/" + from + "/" + to + "/" + strconv.FormatFloat(amount, 'f', 2, 64)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var exchange ExchangeResponse
	err = json.Unmarshal(body, &exchange)
	if err != nil {
		return 0, err
	}

	if exchange.Result != "success" {
		return 0, fmt.Errorf("conversia a eșuat")
	}

	return exchange.ConversionResult, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.step == 0 {
				m.amount, m.err = strconv.ParseFloat(strings.TrimSpace(m.amountStr), 64)
				if m.err == nil {
					m.step = 1
				}
			} else if m.step == 1 {
				m.step = 2
			} else if m.step == 2 {
				m.result, m.err = convert(m.amount, currencies[m.fromIndex], currencies[m.toIndex])
				m.showResult = true
			}

		case tea.KeyBackspace, tea.KeyDelete:
			if m.step == 0 && len(m.amountStr) > 0 {
				m.amountStr = m.amountStr[:len(m.amountStr)-1]
			}

		case tea.KeyUp:
			if m.step == 1 && m.fromIndex > 0 {
				m.fromIndex--
			} else if m.step == 2 && m.toIndex > 0 {
				m.toIndex--
			}

		case tea.KeyDown:
			if m.step == 1 && m.fromIndex < len(currencies)-1 {
				m.fromIndex++
			} else if m.step == 2 && m.toIndex < len(currencies)-1 {
				m.toIndex++
			}

		case tea.KeyRunes:
			if msg.String() == "q" {
				return m, tea.Quit
			}
			if m.step == 0 {
				m.amountStr += msg.String()
			}

		case tea.KeyCtrlC:
			return m, tea.Quit

		default:
			if m.step == 0 && msg.Type == tea.KeyRunes {
				if msg.String() == "q" {
					return m, tea.Quit
				}
				m.amountStr += msg.String()
			}
		}

	}

	return m, nil
}

func (m model) View() string {
	if m.showResult {
		return fmt.Sprintf(
			"\nRezultat: %.2f %s = %.2f %s\n\nApasă q pentru a ieși.",
			m.amount, currencies[m.fromIndex],
			m.result, currencies[m.toIndex],
		)
	}

	switch m.step {
	case 0:
		return fmt.Sprintf("Introdu suma de bani:\n> %s", m.amountStr)
	case 1:
		s := "\nMoneda ta:\n"
		for i, c := range currencies {
			cursor := " "
			if i == m.fromIndex {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, c)
		}
		s += "\nApasă Enter pentru a confirma."
		return s
	case 2:
		s := "\nMoneda în care dorești să schimbi:\n"
		for i, c := range currencies {
			cursor := " "
			if i == m.toIndex {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, c)
		}
		s += "\nApasă Enter pentru rezultat."
		return s
	}
	return ""
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Eroare:", err)
	}
}
