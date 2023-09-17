# door-control
A door controller using [aMember Pro](https://www.amember.com/) and [esprfid](https://github.com/esprfid/esp-rfid).

## Setup

1. Enable the "webhooks" and "Rest API Module" addons in aMember Pro
2. Create a new user field in aMember Pro with the following (minimum) details:
   * Field Name: fob
   * Field Type: SQL
   * SQL field type: Text (string data)
   * Display Type: text
3. Create a new "Remote API key" with the following permissions:
   * Users: index, get
   * Access: index, get
   * Products: index, get
   * Product Category: index, get
   * Product-Product Category: index, get
4. Put all of your membership products under a common category in aMember Pro
5. Create webhooks for `accessAfterInsert`, `accessAfterDelete`, `accessAfterUpdate`, `userAfterInsert`, 
   `userAfterUpdate`, `userAfterDelete` to `https://YOUR.DOOR/v1/amember`
6. Set up your MQTT server and esprfid board(s)

If aMember Pro has a user field named `fob_access`, door-control will make use of the following values:
* `subscription` - door access is tied to aMember Pro access records
* `disabled` - door access is disabled, regardless of aMember Pro records
* `enabled` - door access is enabled, regardless of aMember Pro records

## Installation

Set your environment variables of interest:

```bash
```

Then run `./controller.exe` (or whatever the executable's name is). Downloads are available from the github releases.
