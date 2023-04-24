def disable_download_bar():
    import transformers

    transformers.utils.logging.disable_progress_bar()


disable_download_bar()
