package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lgrees/resy-cli/internal/book"
	"github.com/lgrees/resy-cli/internal/utils/paths"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func sanitizeFilename(name string) string {
	// strip control chars and illegal Windows characters
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	s := re.ReplaceAllString(name, "-")

	// Trim trailing spaces and dots (Windows won't accept these)
	s = strings.TrimRight(s, " .")

	// Avoid reserved device names (CON, PRN, AUX, NUL, COM1..COM9, LPT1..LPT9)
	upper := strings.ToUpper(s)
	reserved := map[string]struct{}{
		"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
	}
	for i := 1; i <= 9; i++ {
		reserved[fmt.Sprintf("COM%d", i)] = struct{}{}
		reserved[fmt.Sprintf("LPT%d", i)] = struct{}{}
	}
	if _, ok := reserved[upper]; ok {
		s = "_" + s
	}

	// avoid ridiculously long names (optional safety)
	if len(s) > 200 {
		s = s[:200]
	}
	if s == "" {
		s = "log"
	}
	return s
}

var bookCmd = &cobra.Command{
	Use:   "book",
	Short: "(internal) Books a reservation immediately",
	Long: `
	Books a reservation using the resy API. This command exists for internal use.
	Generally, users of resy-cli should schedule a booking using "resy schedule".
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		venueId, _ := flags.GetString("venueId")
		partySize, _ := flags.GetString("partySize")
		reservationDate, _ := flags.GetString("reservationDate")
		bookingDateTime, _ := flags.GetString("bookingDateTime")
		reservationTimes, _ := flags.GetStringSlice("reservationTimes")
		reservationTypes, _ := flags.GetStringSlice("reservationTypes")
		dryRun, _ := flags.GetBool("dryRun")

		bookingDetails := &book.BookingDetails{
			VenueId:          venueId,
			PartySize:        partySize,
			BookingDateTime:  bookingDateTime,
			ReservationDate:  reservationDate,
			ReservationTimes: reservationTimes,
			ReservationTypes: reservationTypes,
		}

		p, err := paths.GetAppPaths()
		if err != nil {
			return err
		}
		logDir := p.LogPath
		if logDir == "" {
			// fallback to current working directory if app path is missing
			cwd, _ := os.Getwd()
			logDir = cwd
		}

		venueDetails, _ := book.FetchVenueDetails(venueId)
		venueName := "unknown-venue"
		if venueDetails != nil && venueDetails.Name != "" {
			venueName = venueDetails.Name
		}

		formattedTime := time.Now().Format("Mon Jan _2 15-04-05 2006") // use hyphens instead of colons for filenames
		fileName := sanitizeFilename(fmt.Sprintf("%s_%s.log", venueName, formattedTime))
		fullLogFileName := filepath.Join(logDir, fileName)

		// Create directory (directory only!)
		if err := os.MkdirAll(filepath.Dir(fullLogFileName), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "warning: unable to create log directory %s: %v\n", filepath.Dir(fullLogFileName), err)
		}

		// Try to open the log file. If it fails, fall back to stderr and continue.
		var logWriter io.Writer
		var logFile *os.File
		logFile, err = os.OpenFile(fullLogFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o664)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not open log file %s: %v; falling back to stderr\n", fullLogFileName, err)
			logWriter = os.Stderr
		} else {
			logWriter = logFile
			// only close if we actually opened the file
			defer func() {
				_ = logFile.Close()
			}()
		}

		// create logger on provided writer (never nil)
		l := zerolog.New(logWriter).With().Timestamp().Logger()

		l.Info().Object("booking_details", bookingDetails).Msg("starting book job")
		if bookingDateTime != "" {
			err = book.WaitThenBook(bookingDetails, dryRun, l)
		} else {
			err = book.Book(bookingDetails, dryRun, l)
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(bookCmd)

	flags := bookCmd.Flags()

	flags.String("venueId", "", "The venue id of the restaurant")
	flags.Bool("dryRun", false, "When true, skips booking")
	flags.String("partySize", "", "The party size for the reservation")
	flags.String("bookingDateTime", "", "The time when the reservation should be booked")
	flags.String("reservationDate", "", "The date of the reservation")
	flags.StringSlice("reservationTimes", make([]string, 0), "The times for the reservation")
	flags.StringSlice("reservationTypes", make([]string, 0), "The table types for the reservation")

	bookCmd.MarkFlagRequired("venueId")
	bookCmd.MarkFlagRequired("partySize")
	bookCmd.MarkFlagRequired("reservationDate")
	bookCmd.MarkFlagRequired("reservationTimes")
}
