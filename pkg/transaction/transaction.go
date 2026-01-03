package transaction

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rodrigoTcarmo/inkion/pkg/mail"
	"github.com/rodrigoTcarmo/inkion/pkg/models"
)

type Transaction struct {
	Direction models.Direction
	Amount    float64
	DateTime  *time.Time
}

// ParseTransaction extracts transaction data from an email body
func ParseTransaction(body string) (*Transaction, error) {
	tx := &Transaction{}

	// Extract amount after "Valor recebido:" or "Valor enviado:"
	// Pattern matches: "Valor recebido:\nR$ 1.511,10" or similar variations
	amountRegex := regexp.MustCompile(`(?i)Valor (recebido|enviado):\s*\n?\s*R\$\s*([\d.,]+)`)
	amountMatch := amountRegex.FindStringSubmatch(body)
	if len(amountMatch) > 2 {
		// Convert Brazilian format (1.511,10) to float
		amountStr := amountMatch[2]
		amountStr = strings.ReplaceAll(amountStr, ".", "")  // Remove thousand separator
		amountStr = strings.ReplaceAll(amountStr, ",", ".") // Convert decimal separator
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err == nil {
			tx.Amount = amount
		}
	}

	// Extract date/time: "01 JAN às 23:09"
	dateTimeRegex := regexp.MustCompile(`(\d{2})\s+([A-Z]{3})\s+às\s+(\d{2}:\d{2})`)
	dateTimeMatch := dateTimeRegex.FindStringSubmatch(body)
	if len(dateTimeMatch) > 3 {
		day := dateTimeMatch[1]
		month := dateTimeMatch[2]
		timeStr := dateTimeMatch[3]

		// Map Portuguese month abbreviations to month numbers
		monthMap := map[string]string{
			"JAN": "01", "FEV": "02", "MAR": "03", "ABR": "04", // todo: replace these to constants
			"MAI": "05", "JUN": "06", "JUL": "07", "AGO": "08",
			"SET": "09", "OUT": "10", "NOV": "11", "DEZ": "12",
		}

		monthNum, ok := monthMap[month]
		if ok {
			// Use current year since email doesn't specify
			year := time.Now().Year()
			dateStr := fmt.Sprintf("%d-%s-%s %s", year, monthNum, day, timeStr)
			parsedTime, err := time.Parse("2006-01-02 15:04", dateStr)
			if err == nil {
				tx.DateTime = &parsedTime
			}
		}
	}

	// Determine direction based on email content (you can customize this)
	if strings.Contains(strings.ToLower(body), "valor recebido:") { // todo: replace these strings with custom variables as constants
		tx.Direction = models.DirectionReceived
	} else if strings.Contains(strings.ToLower(body), "valor enviado:") {
		tx.Direction = models.DirectionSent
	}

	return tx, nil
}

func BuildTransaction(emailClient *mail.Mail) ([]*Transaction, error) {
	emails, err := emailClient.FetchEmails()
	if err != nil {
		return nil, err
	}

	var transactions []*Transaction

	for _, email := range emails {
		tx, err := ParseTransaction(email.Body)
		if err != nil {
			continue
		}
		// Only add if valid data is found
		if tx.Amount > 0 || tx.DateTime != nil {
			transactions = append(transactions, tx)
		}
	}

	return transactions, nil
}
