# Unleash Service

This service allows to interact with the open source feature toggle system [unleash](https://github.com/unleash). 
Triggered by a cloudevent with 

## Local Development

If you want to locally develop, test and run the service, there are some steps to ease this:

- Set up port-forwarding to your Unleash server:

    ```
    kubectl port-forward svc/unleash-server-service 7000:80 
    ```

- Set environment variable for your unleash server `UNLEASH_SERVER_URL`, e.g., in your VSCode launch.json: 

    ```
    {
        // Use IntelliSense to learn about possible attributes.
        // Hover to view descriptions of existing attributes.
        // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
        "version": "0.2.0",
        "configurations": [
            {
                "name": "Launch",
                "type": "go",
                "request": "launch",
                "mode": "auto",
                "program": "${fileDirname}",
                "env": { "UNLEASH_SERVER_URL":"http://localhost:7000" },
                "args": []
            }
        ]
    }
    ```

