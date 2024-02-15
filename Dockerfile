FROM golang:alpine as build

# ---- ENV ----
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

ARG SERVICE_NAME
ENV SERVICE_NAME=$SERVICE_NAME
ENV PROJECT_PATH=job_interview_appointment_api

RUN apk add --no-cache make git tzdata

WORKDIR /go/src/${PROJECT_PATH}
ADD . ./

RUN rm -f go.mod
RUN rm -f go.sum
RUN rm -rf ./dist
RUN go mod init
RUN go mod tidy

RUN go build -o ./dist/${SERVICE_NAME}

# ----------------------
# ALPLINE IMAGE
# ----------------------

FROM alpine

ARG SERVICE_NAME
ENV SERVICE_NAME=$SERVICE_NAME
ENV PROJECT_PATH=job_interview_appointment_api

ARG APP_ENV
ENV APP_ENV=$APP_ENV

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /go/src/${PROJECT_PATH}

COPY .env.${APP_ENV} ./.env.${APP_ENV}
COPY rbac_model.conf ./rbac_model.conf
COPY --from=build /go/src/${PROJECT_PATH}/dist/${SERVICE_NAME} ./

RUN echo "./${SERVICE_NAME} -env \$APP_ENV" >>./endpoint.sh
RUN chmod 777 ./endpoint.sh

CMD ./endpoint.sh
