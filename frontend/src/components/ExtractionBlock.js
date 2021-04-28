import React, { useState } from 'react';

const API_URL = 'http://localhost:8080/api';

const ExtractionBlock = () => {
	const [selectedFile, setSelectedFile] = useState(undefined);
	const [fileUploaded, setFileUploaded] = useState(false);

	const uploadImage = (e) => {
		e.preventDefault();

		console.log('Uploading image.');

		const formData = new FormData();
		formData.append('image', selectedFile);

		const options = {
			method: 'POST',
			headers: {
				'Content-Type': 'multipart/form-data',
				'Accept': 'application/json',
			},
			body: formData,
		}
		delete options.headers['Content-Type'];
		fetch(API_URL+'/upload', options)
		.then(response => {
			console.log(response);
			if (response.status === 200) {
				setFileUploaded(true);
			}
		})
		.catch(error => {
			console.log('Upload failed.');
			console.error('Error:', error);
		});
	}

	return (
		<div className="extraction-block">
			<PaletteExtractForm 
				onSubmit={uploadImage} 
				selectedFile={selectedFile}
				setSelectedFile={setSelectedFile}
				fileUploaded={fileUploaded}
			/>
		</div>
	);
}

const PaletteExtractForm = (props) => {
	return (
		<div>
			<form onSubmit={(e) => props.onSubmit(e)}>
				<div>Upload.</div>
				<input 
					type="file" 
					name="image"
					onChange={(e) => props.setSelectedFile(e.target.files[0])}
				/>
				<button>
					Upload image!
				</button>
			</form>
			<DisplayPalette 
				display={props.fileUploaded}
			/>
		</div>
	);
}

const DisplayPalette = ({ display }) => {
	const [numCols, setNumCols] = useState(3);
	const [palette, setPalette] = useState([]);

	if (!display) {
		return null;
	}

	const getPalette = () => {
		const urlToUse = `${API_URL}/extract/?colors=${numCols}`;

		const options = {
			method: 'GET',
		};

		fetch(urlToUse, options)
			.then(response => response.json())
			.then(result => {
				console.log('Returned:', result);
				setPalette(result);
			})
			.catch(error => {
				console.log('Palette extraction failed.');
				console.error('Error:', error);
			});
	}

	const options = [];
	for (let i = 3; i < 11; i++) {
		options.push(<option value={i} key={i}>{i}</option>);
	}

	const blocks = [];
	for (let i = 0; i < palette.length; i++) {
		blocks.push(
			<div 
				className="color-block" 
				style={{backgroundColor: palette[i].color}}
				key={i}
			>
				{palette[i].color}: {palette[i].frequency}
			</div>
	);
	}

	return (
		<div>
			<div>
				<label htmlFor="extractNum">Colors to extract:</label>
				<select 
					id="extractNum"
					value={numCols}
					onChange={(e) => setNumCols(e.target.value)}
				>
					{options}
				</select>
				<button onClick={getPalette}>
					Find palette!
				</button>
			</div>
			<div className="display-palette">
				{
					blocks
				}
			</div>
		</div>
	);
}

export default ExtractionBlock;
