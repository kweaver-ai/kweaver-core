# Python Data Science Template
# Python 3.11 environment with data science libraries
# Based on sandbox-executor-base which contains the executor daemon

ARG BASE_IMAGE=sandbox-executor-base:latest
FROM ${BASE_IMAGE}

# Metadata
LABEL template.name="python-datascience"
LABEL template.description="Python 3.11 with data science libraries (pandas, numpy, matplotlib)"
LABEL template.version="1.0.0"

# Switch to root to install packages
USER root

# Install system dependencies for data science libraries
RUN apt-get update && apt-get install -y \
    # Build dependencies
    build-essential \
    # Plotting dependencies
    libfreetype6-dev \
    libpng-dev \
    libjpeg-dev \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Install Python data science packages
RUN pip install --no-cache-dir \
    pandas==2.1.4 \
    numpy==1.26.2 \
    matplotlib==3.8.2 \
    scipy==1.11.4 \
    scikit-learn==1.3.2 \
    seaborn==0.13.0 \
    jupyter==1.0.0

# Switch back to sandbox user
USER sandbox

# CMD is inherited from the base image (executor daemon)
