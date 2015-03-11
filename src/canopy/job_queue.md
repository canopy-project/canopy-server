

For scalability, we will restructure Canopy to use a job queue.  Whenever
something needs to happen, a job request will be posted to the queue and picked
up by a worker thread.

Some job requests must be fullfilled by a particular worker.  For example, some
workers have open websocket connections to a device.

Originators.
New jobs will get created when.

 - REST API request
    - Each REST API request immediately turns into a RESTJob and gets forwarded
      to some worker.

 - Websocket request
    - Each received websocket payload turns into a WSJob and gets forwarded to
      some worker.

 - Pigeon Relay (REST API -> Websocket)
    - Whenever a cloud variable changes, it must be forwarded to the
      appropriate server/thread.

 - Rule Input change
    - Whenever a cloud variable changes, any rules that depend on the cloud
      variable must be re-evaluated.


