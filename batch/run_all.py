"""Main entry-point for the data-platform-batch service.

Runs:
  1. fetch_ohlcv.run_all()  — every FETCH_INTERVAL_MINUTES (default 60)
  2. compute_indicators.run_all() — COMPUTE_INTERVAL_MINUTES later (default 65)

Also starts a lightweight gRPC server implementing the BatchData proto service.
"""
import logging
import threading

import schedule

import config
import fetch_ohlcv
import compute_indicators
from database import init_db
from batch_grpc_server import serve_grpc

logger = logging.getLogger(__name__)


def _run_pipeline() -> None:
    logger.info("pipeline: fetch OHLCV start")
    fetch_ohlcv.run_all()
    logger.info("pipeline: compute indicators start")
    compute_indicators.run_all()
    logger.info("pipeline: done")


def main() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")

    logger.info("initialising database tables")
    init_db()

    # Run once on startup
    _run_pipeline()

    # Schedule recurring runs
    schedule.every(config.FETCH_INTERVAL_MINUTES).minutes.do(fetch_ohlcv.run_all)
    schedule.every(config.COMPUTE_INTERVAL_MINUTES).minutes.do(compute_indicators.run_all)

    # Start gRPC server in background thread
    grpc_thread = threading.Thread(target=serve_grpc, args=(config.GRPC_PORT,), daemon=True)
    grpc_thread.start()
    logger.info("gRPC server started on port %d", config.GRPC_PORT)

    # Run scheduler loop (blocks)
    import time
    while True:
        schedule.run_pending()
        time.sleep(10)


if __name__ == "__main__":
    main()
