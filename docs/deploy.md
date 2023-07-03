## Deploy into kubernetes cluster

`Prepare Config Map`


```
cat  << EOF > app-config.yaml
port: 5115
endPoints:
- name: google
  url: https://www.google.com
- name: frontend
  url: http://frontend.magellan.svc.cluster.local:8080/ping
- name: backend
  url: http://backend.magellan.svc.cluster.local:8081/ping
- name: storage 
  url: http://storage.magellan.svc.cluster.local:8082/ping
EOF
```

That command creates a yaml file, `app-config.yaml`

**create config map**

`kubectl create configmap app-cm --from-file=app-config.yaml -n magellan`


## Create a secrets for docker registry
```shell
kubectl create secret docker-registry samutup-secrets \
--docker-server=https://hub.docker.com \
--docker-username=samutup \
--docker-password=Balinese100% \
--namespace magellan \
--output yaml --dry-run=client | kubectl apply -f -
```

## Create a service account that holds the `imagePullSecrets`

```shell
kubectl create serviceaccount netpol-sa -n magellan
```

`Patch the service account to link it to the imagePullSecrets`

```shell
kubectl patch serviceaccount -n magellan netpol-sa \
-p "{\"imagePullSecrets\": [{\"name\": \"samutup-secrets\" }] }"
```
<a id="deployment_manifest">Deployment manifest</a>
```shell
kubectl apply -n magellan -f - << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: frontend
  name: frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
  strategy: {}
  template:
    metadata:
      labels:
        app: frontend
    spec:
      serviceAccountName: netpol-sa
      containers:
      - image: samutup/http-ping:0.0.6
        name: http-ping
        env:
        - name: APP_NAME
          value: Frontend
        command:
        - "/app/http-ping"
        args:
        - "launchHttp"
        - "--appName=frontend"
        - "--config=/app/config/app-config.yaml"
        securityContext:
          allowPrivilegeEscalation: false
          runAsNonRoot: true
          runAsUser: 1000
          readOnlyRootFilesystem: true
        resources: {}
        volumeMounts:
        - mountPath: /app/config
          name: fe-v
        readinessProbe:
          httpGet:
            path: /ping
            port: 5115
          periodSeconds: 10
          initialDelaySeconds: 5
          failureThreshold: 5
          successThreshold: 1
      volumes:
      - name: fe-v
        configMap:
          name: app-cm
          items:
          - key: app-config.yaml
            path: app-config.yaml
EOF
```

Verify if the deployment is working fine.


`k get all -n magellan -l app=frontend`


Expose the deployment into service,


`kubectl expose deployment -n magellan frontend --port 8080 --target-port 5115`

Verify if service created,

`k get all -n magellan -l app=frontend`

```shell
NAME                            READY   STATUS    RESTARTS   AGE
pod/frontend-86b7fb7dc7-gtb8z   1/1     Running   0          14m

NAME               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/frontend   ClusterIP   10.101.93.214   <none>        8080/TCP   35s

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/frontend   1/1     1            1           14m

NAME                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/frontend-86b7fb7dc7   1         1         1       14m
```