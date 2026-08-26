FROM eclipse-temurin:26-jdk AS toolchain

# renovate: datasource=github-releases depName=JetBrains/kotlin
ARG KOTLIN_VERSION=2.4.10

ADD https://github.com/JetBrains/kotlin/releases/download/v${KOTLIN_VERSION}/kotlin-compiler-${KOTLIN_VERSION}.zip /tmp/kotlin-compiler.zip

RUN cd /opt && jar xf /tmp/kotlin-compiler.zip && chmod +x /opt/kotlinc/bin/*

FROM eclipse-temurin:26-jdk

COPY --from=toolchain /opt/kotlinc /opt/kotlinc

ENV PATH="/opt/kotlinc/bin:${PATH}"
