from setuptools import setup, find_packages

setup(
    name="plf-py",
    version="1.0.0",
    author="antnet1094",
    description="Python bindings for Prompt Language Format (PLF)",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    url="https://github.com/antnet1094/plf",
    packages=find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires='>=3.8',
    install_requires=[], # No tiene dependencias, solo el binario de Go
)
