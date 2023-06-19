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

## How it works

This is a command-line tool first of all. Tests will be run just once and script will be finished.
TODO describe options 

## How to write tests

TODO define initial steps

Here is the detailed template of a test case:
```yaml
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
    # TODO describe store and variables
    # 
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
        age: { gte: 13 }
        name: "some name"
        created: { gt: 1254568 }
      some_field: 5
```

Additional services can be added in `services.yaml` file:
```yaml
{SERVICE_ALIAS}:
    service: {PROTO_SERVICE}
    address: "{HOST}:{PORT}"
    # metadata that will be provided in each request (optional)
    metadata:
      {KEY}: {VALUE}
```

TODO global config