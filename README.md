## What and Why

This service was created to make grpc calls against microservices in order to
test them itself as long as integration with their dependencies in live environment.
To create test, you just need describe your case in configuration file, without code.

So, its kinda hybrid of functional and integration testing, because it would be too
expensive to run functional tests as part of CI process for microservices, which can
have numerous of dependencies as biggest part of their functionality, so mocking could
take most of the time, which you can spend more efficiently. On the other hand,
setup all the dependencies for each microservice in container to run integration tests against
target app would be too complex and expensive from time (and, perhaps, resources) perspective.

## Glossary

* Test case - instructions, that describe some specific case of using specific method of
  specific service. Divided into steps.
* Step - one call to specific method of some service that describes input and expected output,
  if needed. May not test anything on its own, but only be part of a multi-stage test.
* Run - database entity, that contains information about single run of this service in command-line or
  worker modes.

## Installation

Download
```bash
wget https://github.com/Gregmus2/grpc-fts/releases/download/v1.4.0/fts && sudo chmod +x fts
```
Run init command to create a new project in current directory
```shell
./fts init --directory .
```

It will create a new project with the following structure:
```
test-cases/     <- put your test cases here
    .gkeep  
global.yaml     <- basic required configuration
services.yaml   <- services definition
variables.yaml  <- variables definition if needed
```

Now you can fill generated configuration, write test cases and then validate them.
```shell
./fts validate
```

If everything is ok, you can run tests
```shell
./fts run
```

## How it works

This is a command-line tool first of all. Tests will be run just once and script will be finished.

## How to write tests

Here is the detailed template of a test case:
```yaml
# in case if your test has any dependencies, you can describe them here
depends_on:
  - init # name of another test case without .yaml extension

# in some cases you will need to run several steps 
#   to provide expected pre-requirements for your testing.
# they will be run in order they were described in this test case file
steps:
    # service is alias from services.yaml file which reference to 
    #   some specific microservice of your product (foo in this case).
    # services.yaml will be described later in this README file
  - service: foo
    # method from proto files of the target service
    method: GetUserData
    # in the request you should describe proto request as it would be
    #   written in json. you can get request fields from proto files.
    request:
      group_id: "some id"
      filters:
        - name: "some name"
          value: "some value"
    # here, you can specify expected response for this call. 
    # all fields are optional, you need to describe only fields and conditions
    #   that you want to check
    #
    # available methods:
    #     len    - to check length of the array ( assets: { len: 10 } )
    #     gt     - greater than ( bedrooms: { gt: 1 } )
    #     gte    - greater than or equal ( bedrooms: { gte: 1 } )
    #     lt     - lesser than  ( bedrooms: { lt: 3 } )
    #     lte    - lesser than or equal  ( bedrooms: { lte: 2 } )
    #     not    - condition or value ahead is not true ( not: 2 )
    #     first  - check first value of a slice ( some_array: { first: 2 } )
    #     one_of - value expected to be equal at least with one of elements
    #         ( Country: { one_of: [IT, CY, GR] } )
    #         ( one_of:
    #             - Latitude: 52.5
    #               Longitude: 25.5
    #             - Latitude: 65.8
    #               Longitude: 14.1
    #             - Latitude: 52.5
    #               Longitude: 45.6 )
    #      any - at least one of elements of target array should satisfy the condition.
    #         embedded conditions are allowed.
    #         ( any:
    #             prediction: { gt: 3 }
    #             id: some id )
    #      all - all elements of target array should satisfy the condition
    #         embedded conditions are allowed.
    #         ( all: prediction: { gt: 3 } )
    #      store - store value to use it in another step
    #         ( foo_property: { store: fooVariable } )
    #         You will be able to use this value in another step as $fooVariable.
    #         You can use them both in request and response.
    #         You can use variables from variables.yaml or command option in the same way
    #      
    
    #      Also you have an option to use full slice match to check if all elements of target array are present 
    #      in the expected array. Simply - order independent full match of arrays.
    #         embedded conditions are allowed.
    #         ( counts:
    #             - poi_type: Culture
    #               count: 1
    #             - poi_type: Food
    #               count: 7
    #             - poi_type: Education
    #               count: 3 )
    
    response:
      user_data:
        id: { store: $userID } # <- I can store response value to use it in another step
        age: { gte: 13 }
        name: "some name"
        created: { gt: 1254568 }
      some_field: 5
    # metadata that will be provided in each request (optional)
    metadata:
      # You can specify timeout for this step. A duration string is a possibly signed sequence of
      #   decimal numbers, each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or "2h45m".
      #   Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
      timeout: 5s
      var1: "value1"
      
  - service: bar
    method: SendEmail
    request:
      user_id: $userID # <- I can use stored value from previous step
      message: "some message"
    response:
      status: "OK"  
...
```

Additional services can be added in `services.yaml` file:
```yaml
{SERVICE_ALIAS}:
    service: {PROTO_SERVICE}
    address: "{HOST}:{PORT}"
    # metadata that will be provided in each request (optional)
    metadata:
      {KEY}: {VALUE}
...
```

Global config:
```yaml
proto_root: "path to root directory with your proto files"
proto_sources:
  - "it can be relative path (in proto root) to directory with proto files"
  - "or it can be relative path (in proto root) to some specific proto file"
proto_imports:
  - "path to additional proto imports, like google protobuf utilities for example"
```

Variables:
```yaml
{VARIABLE_NAME}: {VALUE}
...
```

## Protobuf

Currently, most of proto mechanisms are supported. You can use enums, messages, oneof, maps, etc.
If you are not sure how to describe some type in test case, think of it as if it would be json request or response,
which you would send or receive using most of grpc clients.

### Streams

Currently only server side streams are supported. Within this mode, the response should contain "stream" key 
with array of elements. In order to prevent utility from stuck, you can specify "timeout" key under "metadata" key in step object. 
It will be applied for entire call (including all messages).
Status checks will apply for each element of the stream.

Example
```yaml
steps:
  - service: foo
    method: Listen
    request:
      id: "some id"
    metadata:
      timeout: 5s
    response:
      stream:
        - user_data:
            age: { gte: 13 }
            name: "some name"
            created: { gt: 1254568 }
      some_field: 5
```

## Troubleshooting

### field XXX is not function, neither field

Try to write response fields in CamelCase