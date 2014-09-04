# Imsto Api

## Defines
- All writable api use POST mothod
- All readonly api use GET mothod
- The base64 encoding is url safety

### roof
- type: string
- Section name that in config 'imsto.ini'
- All apis must include this argument

### api_key
- type: string
- Api caller key, for api authorization
- All apis must include this argument

### user
- type: uint
- User ID

### token
- A hashed string for authorization
- urlencoding with base64

### ticket
- Same as a token, but include a extensional ticket ID
- For upload in mobile or other third environment
- struct:
  - appid: int
  - author: int
  - prompt: string,
  - img_id: string, uploaded image ID
  - img_path: string, upload image path without prefix
  - done: boolean, has finished?


## Apis list
- All api response struct:
  - `status`: string `ok` or others
  - `data`: array or object, result
  - `error`: string, error messages, just only `status !== "ok"`

### Request a new token
- method: `POST /imsto/token`
- args: `roof,api_key,user`

### Request a new ticket
- method: `POST /imsto/ticket`
- args: `roof,api_key,user,token`

### Check a ticket
- method: `GET /imsto/ticket`
- args: `roof,api_key,token`
- note: the token must be a Ticket Token

### Upload files
- method: `POST /imsto/`
- content type: `multipart/form-data`
- args: `roof,api_key,user,token,file`
- note: 1. input name must use `file`; 2. the token must be a Ticket Token


## Mobile upload workflow

1. Request new ticket and drop a barcode in web browser
> `POST /imsto/ticket`
2. Take the barcode that include ticket token in iOS program
3. Check the ticket is valid and unfinished, read it's prompt for display
4. Post one or more files to server with argument token
> `POST /imsto/`
> mutilpart-data: file=FILE,roof=ROOF,token=TICKET_TOKEN
