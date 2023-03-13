def load_notepbook(notebook_path):
    """
    Read the jupyter notebook as json.
    """
    import json

    with open(notebook_path, "r") as f:
        notebook = json.load(f)
    return notebook


def write_markdown(notebook, output_path):
    """
    Loop through all the cells in the JSON notebook and write to
    write them as markdown.
    """
    with open(output_path, "w") as f:
        for cell in notebook["cells"]:
            write_cell(cell, f)


def write_cell(cell, f):
    """
    Write a single cell as markdown based on the cell type.
    This currently only supports "mardown" and "code" cells.
    """
    f.write("\n\n\n<!-- ------------- New Cell ------------ -->\n\n\n")
    if cell["cell_type"] == "markdown":
        write_markdown_cell(cell, f)
    elif cell["cell_type"] == "code":
        write_code_cell(cell, f)
    else:
        print("Unknown cell type:", cell["cell_type"])
    f.write("\n\n")


def write_markdown_cell(cell, f):
    """
    Write markdown cell as is.
    """
    f.writelines(cell["source"])


def write_code_cell(cell, f):
    """
    Write code cell if there is code. It may be an empty cell in which case
    nothing is written.  If code is found it is assumed to be python.
    If the code has output that is written here as well.
    """
    if "".join(cell["source"]).strip():  # drop empty source cells
        f.write("```python\n")
        f.writelines(cell["source"])
        f.write("\n```")
        if "outputs" in cell:
            write_outputs(cell["outputs"], f)


def write_outputs(outputs, f):
    """
    Write output for a code cell.  This currently only supports
    "text/html" and "text/plain" output types.
    """
    for output in outputs:
        if output["output_type"] == "execute_result":
            if "text/html" in output["data"]:
                f.write("\n**Output**\n")
                html_lines = "".join(output["data"]["text/html"])
                import re

                html_lines_without_style = re.sub(r"<style(.|\n)*</style>", "", html_lines)

                # This line is needed because scikit-learn has started
                # detecting whether the user is running in a Jupyter notebook
                # and prints out some silly HTML box for its models when the
                # user is in Jupyter. This renders some spammy message on our
                # docs, so we hardcode a rule to get rid of it.
                html_lines_without_skl_warn = re.sub(
                    r"<b>In a Jupyter environment(.|\n)*</b>", "", html_lines_without_style
                )

                f.writelines(html_lines_without_skl_warn)
            elif "text/plain" in output["data"]:
                f.write("\n**Output:**\n")
                f.write("\n\n```\n")
                f.writelines(output["data"]["text/plain"])
                f.write("\n```\n\n")


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--input", required=True, help="The relative path to the notebook to render."
    )
    parser.add_argument(
        "--output", required=True, help="The relative path to the desired markdown location."
    )
    args = parser.parse_args()
    notebook = load_notepbook(args.input)
    write_markdown(notebook, args.output)
    print("Notebook rendered to markdown successfully!")
