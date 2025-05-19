FROM ubuntu:24.04

# Install system dependencies

# Install prerequisites
RUN apt-get update && apt-get install -y \
    software-properties-common \
    build-essential \
    curl \
    wget \
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libffi-dev \
    libncursesw5-dev \
    xz-utils \
    tk-dev \
    git \
    liblzma-dev \
    lzma \
    ca-certificates \
    libmariadb-dev \
    gcc

# Add deadsnakes PPA and install Python 3.11
RUN add-apt-repository ppa:deadsnakes/ppa && apt-get update && apt-get install -y python3.11 python3.11-venv python3.11-dev

# Create a venv with Python 3.11
RUN python3.11 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

WORKDIR /app

# Activate the virtual environment and install Python dependencies
# Copy slackbot-specific requirements.txt from the 'slackbot' directory in the build context
COPY slackbot/requirements.txt .
RUN . /opt/venv/bin/activate && pip3 install --no-cache-dir -r requirements.txt

# Copy the protos directory from the repository root (build context) to /app/protos
COPY protos ./protos

# Copy the slackbot application code from the 'slackbot' directory in the build context to /app/slackbot
COPY slackbot ./slackbot

# Command to run the application using the virtual environment's Python
# The main script is program.py inside the /app/slackbot directory
CMD ["/opt/venv/bin/python3", "slackbot/program.py"]
