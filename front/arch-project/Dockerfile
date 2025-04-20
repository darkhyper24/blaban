# Build stage only - no hosting container needed
FROM node:23-alpine

WORKDIR /app

COPY package*.json ./

RUN npm ci

COPY . .

RUN npm run build

# Create a volume mount point for the build output
VOLUME /app/dist

# Keep container running only during development if needed
# In production, this container only needs to run during the build phase
CMD ["tail", "-f", "/dev/null"]