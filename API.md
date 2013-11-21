# Imsto Api

## Defines
- All writable api use POST mothod
- All readonly api use GET mothod
- The base64 encoding is url safety

### roof
- type: string
- Section name that config 'imsto.ini'
- All apis must include this argument

### app
- type: uint8
- Custom app ID
- All apis must include this argument

### user
- type: uint8
- User ID

### token
- A hashed string for authorization
- urlencoding with base64

### ticket
- Same as a token, but include a extensional ticket ID
- For mobile or other third environment upload


## Apis list

### Request a new token
- method: `POST /imsto/token`
- args: `roof,app,user`

### Request a new ticket
- method: `POST /imsto/ticket`
- args: `roof,app,user,token`

### Check a ticket
- method: `GET /imsto/ticket`
- args: `roof,app,token`
- note: the token must be a Ticket Token

### Upload files
- method: `POST /imsto/`
- content type: `multipart/form-data`
- args: `roof,app,user,token,file`
- note: 1. input name must use `file`; 2. the token must be a Ticket Token


## Mobile upload workflow

1. Request new ticket and drop a barcode in web browser
> `POST /imsto/ticket`
2. Take the barcode that include ticket token in iOS program
3. Post one or more files to server with argument token
> `POST /imsto/`
> mutilpart-data: file=FILE,roof=ROOF,token=TICKET_TOKEN