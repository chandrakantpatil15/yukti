# Multi-stage build with Maven cache
FROM alpine:3.18 as builder
WORKDIR /app

# Install OpenJDK 17 and Maven
RUN apk add --no-cache openjdk17 maven

# Copy pom.xml first for dependency caching
COPY pom.xml .
RUN mvn dependency:go-offline -B

# Copy source and build
COPY src ./src
RUN mvn clean package -DskipTests

FROM alpine:3.18
WORKDIR /app

# Install OpenJDK 17 JRE
RUN apk add --no-cache openjdk17-jre

COPY --from=builder /app/target/*.jar app.jar
EXPOSE 8090
ENTRYPOINT ["java", "-jar", "app.jar"]