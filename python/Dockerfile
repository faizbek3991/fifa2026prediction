FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
RUN mkdir -p /data
ENV DB_PATH=/data/users.db
ENV PORT=8501
EXPOSE 8501
# Shell form so $PORT expands -- Render (and similar hosts) assign this at runtime.
ENTRYPOINT streamlit run app.py --server.port=$PORT --server.address=0.0.0.0
