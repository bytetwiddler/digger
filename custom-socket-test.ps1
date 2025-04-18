function Test-CustomTcpConnection {
    param (
        [string]$TargetHost,
        [int]$Port,
        [int]$TimeoutMilliseconds = 5000
    )

    try {
        # Create a new TcpClient instance
        $client = New-Object System.Net.Sockets.TcpClient

        # Begin an asynchronous connection attempt
        $asyncResult = $client.BeginConnect($TargetHost, $Port, $null, $null)

        # Wait for the connection attempt to complete or timeout
        if ($asyncResult.AsyncWaitHandle.WaitOne($TimeoutMilliseconds)) {
            $client.EndConnect($asyncResult) # Complete the connection
            Write-Output "Connection to $TargetHost on port $Port successful."
            $client.Close()
        } else {
            # Timeout occurred
            Write-Output "Connection to $TargetHost on port $Port timed out after $TimeoutMilliseconds milliseconds."
	    Write-Output "No TCP soup for you"        
	}
    } catch {
        # Handle exceptions such as DNS resolution failure or refusal
        Write-Output "Connection to $TargetHost on port $Port failed: $_"
	Write-Output "No TCP soup for you"
    } finally {
        # Clean up resources
        if ($client) {
            $client.Close()
        }
    }
}

$ExternalIP = Invoke-RestMethod -Uri 'http://api.ipify.org'

# Import the CSV file
$csvData = Import-Csv -Path "sites.csv"

foreach ($row in $csvData) {
	$Server = $row.Hostname
	$Port = [int]$row.Port
	# Example usage
	Test-CustomTcpConnection -TargetHost $Server -Port $Port -TimeoutMilliseconds 6000
}

# Display the external IP address
Write-Host "Your external IP address is: $ExternalIP"
