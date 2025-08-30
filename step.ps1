 Write-Host "ENV DEBUG: RES_DAYS_AHEAD='$env:RES_DAYS_AHEAD' VENUE_ID='$env:VENUE_ID' PARTY_SIZE='$env:PARTY_SIZE'"

              # PowerShell on windows-latest
              $ErrorActionPreference = "Stop"

              # compute reservation date
              $date = (Get-Date).AddDays($env:RES_DAYS_AHEAD).ToString('yyyy-MM-dd')
              Write-Host "Reservation date: $date"

              $times = @(
                "17:00:00","17:15:00","17:30:00","17:45:00",
                "18:00:00","18:15:00","18:30:00","18:45:00",
                "19:00:00","19:15:00","19:30:00","19:45:00",
                "20:00:00","20:15:00","20:30:00","20:45:00",
                "21:00:00","21:15:00","21:30:00","21:45:00",
                "22:00:00","22:15:00","22:30:00","22:45:00",
                "23:00:00","23:15:00","23:30:00","23:45:00",
                "00:00:00"
              )
              $env:PARTY_SIZE = 2
              $env:VENUE_ID = 60058
              foreach ($t in $times) {
                echo book  --partySize=$env:PARTY_SIZE      --reservationDate=$date   --reservationTimes=$t   --venueId=$env:VENUE_ID --reservationTypes=""

                Write-Host "Attempting booking at time: $t"
                sudo .\resy-cli-windows-amd64.exe book --partySize=$env:PARTY_SIZE  --reservationDate=$date --reservationTimes=$t --venueId=$env:VENUE_ID --reservationTypes=""              }

              Write-Host "All attempts finished."