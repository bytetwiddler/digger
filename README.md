
# Digger:

Check and report on changes in DNS resolution for a list of sites. 

## Description

### The problem:
If your organization restricts outbound services to sites that do not whitelist client access,
then chances are the third party service does not publish IP address changes.  In addition if
your organization manages outbound access at layer 3-4, then they are using the IP address, protocl 
and ports of the third party service, not the domain name (layer 7). This means that when the IP
address of the third party changes it likely will result in your outbound connection being dropped.

### What does the digger service do to help with the above problem? 
The digger service keeps track of the last known IP address for the thrid party site. It periodically 
checks the DNS resolution of the URI's for your third party sites, it then compares the last known IP 
address against the IP's returned in the DNS response and if it does not find a match it will send an 
email to a configurable email address with the notification and details of what has changed.


## Getting Started

### Build Dependencies

* Go >= 1.24
* Make, Makefile is currently setup to build on windows. Though of course you could build from the command line.

### Runtime Dependencies

* Windows host
* Admin Prvileages for the service_manager application
* SMTP forwarder

### Building it (In Powershell)
  ```
  PS C:\digger> make all
  ```
### Make zip of required files for deployment on windows  
  ```
  PS C:\digger> make zip
  ```

### Installing (In Powershell with elevated priv's)
* Move the zip file to a folder on the Windows system you want to install the service
* Unzip the zip file
* Edit the config.yaml file with the secifics of your installation
* Edit the sites.csv file with minimally with the first 3 fields (Hostname, Port, EntityName,,,,) with the sites you want to monitor
* Run the digger service with the -update flag to populate the current IP addresses for the services found in sites.csv
  ```
  PS C:\digger> .\digger-windows-amd64.exe -update
  ```
* Run the following command to install the service
  ``` 
  PS C:\digger> .\service_manager -install
  ```
* Start the service
  ```
  PS C:\digger> .\service_manager -start
  ```
This should immediately start the digger service, then it should run again every 4 hours.

### Executing the digger service program outside of the service
There are occasions such as the set above when you will want to run the service outside of the windows service intervals.  
You can do that one of 2 ways.  
  1) Stop/Start the service
  2) Run the digger service from powershell
     ```
     PS C:\digger> .\digger-windows-amd64.exe
     PS C:\digger> .\digger-windows-amd64.exe -update
     PS C:\digger> .\digger-windows-amd64.exe -report
     ```
     #### Usage:
     Run digger in normal mode.  Digger will verify the IP address and report changes.
       ```
        .\digger-windows-amd64.exe 
       ```
     Run digger with -update flag. Digger will verify the IP address, report changes, update the sites.csv file with the current IP address for the site.
       ```
        .\digger-windows-amd64.exe -update
       ```
     Run digger with the -report flag. Digger will report previous changes to the IP address, the old IP, the new IP and a timestamp of when it found that change.
       ```
        .\digger-windows-amd64.exe -report
       ```

     Note: for the most part digger does not output to stdout or stderr.  It write to digger.log as configured in the config.yaml.  It also writes windows events.
     The one exception to this is when the '-report' flag is used. It will write that output to stdout.
     
* Step-by-step bullets
```
code blocks for commands
```

## The Future

1) Make time between iterations configurable
2) Move the email template outside of the code
3) Use go plugin interface for notifications.  The current implementation just sends an email, but perhaps you want it to create a ticket in your ticketing system,etc...
4) Add more unit test coverage
5) Fix the myriad of linting issue

## Why go?
Why not? I did this on my own time over a single evening to solve a problem for my current employer. I used the language I wanted to use. 

Seems like it could be replaced with a 5 line script in 'insert your fav scripting language here'?  Perhaps but again see the line above.

This is not idiomatic go. Probably not, I'm always ready to be smarter tomorrow than I am today, I am open to pull requests.

## Why is it named Digger?
Because I'm not very imaginative and use dig all the time for DNS related chores.

## Version History

* 0.0.1
  * Initial Release

## License

This project is licensed under the MIT License - Use at your own peril

## Acknowledgements
* We stand on the shoulders of giants, read the go.mod for a list of those giants
