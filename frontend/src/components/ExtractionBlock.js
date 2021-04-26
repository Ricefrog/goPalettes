import React, { useState } from 'react';

const API_URL = 'http://localhost:8080/api';

const ExtractionBlock = () => {
	const placeholderColor = "#DAF7A6";

	const defaultColor = {
		color: placeholderColor,
		frequency: 0,
	};

	const [numCols, setNumCols] = useState(3);
	const [selectedFile, setSelectedFile] = useState(undefined);
	const [palette, setPalette] = useState(
		[
			defaultColor, 
			defaultColor, 
			defaultColor, 
		]
	);

	const submitForm = (e) => {
		e.preventDefault();

		console.log('Uploading extract form.');
		const formData = new FormData();
		formData.append('image', selectedFile);
		formData.append('numOfColors', numCols);

		const options = {
			method: 'POST',
			mode: 'no-cors',
			headers: {
				'Content-Type': 'multipart/form-data',
			},
			body: formData,
		}
		delete options.headers['Content-Type'];

		fetch(API_URL+'/extract', options)
		.then(response => response.json())
		.then(result => {
			console.log('Success:', result);
			setPalette(result);
		})
		.catch(error => {
			console.error('Error:', error);
		});
	}

	const colHandler = (num) => {
		// change numCols in addition to palette
		setNumCols(num);
		const newPalette = [];
		for (let i = 0; i < num; i++) {
			newPalette.push(defaultColor);
		}
		setPalette(newPalette);
	}

	return (
		<div className="extraction-block">
			<PaletteExtractForm 
				onSubmit={submitForm} 
				numCols={numCols}
				setNumCols={colHandler}
				selectedFile={selectedFile}
				setSelectedFile={setSelectedFile}
			/>
			<DisplayPalette 
				palette={palette}
			/>
		</div>
	);
}

const PaletteExtractForm = (props) => {

	const options = [];
	for (let i = 3; i < 11; i++) {
		options.push(<option value={i} key={i}>{i}</option>);
	}

	return (
		<form onSubmit={(e) => props.onSubmit(e)}>
			<div>Extract a palette from a file.</div>
			<label htmlFor="extractNum">Colors to extract:</label>
			<select 
				id="extractNum"
				value={props.numCols}
				onChange={(e) => props.setNumCols(e.target.value)}
			>
				{options}
			</select>
			<input 
				type="file" 
				name="image"
				onChange={(e) => props.setSelectedFile(e.target.files[0])}
			/>
			<button>
				Find palette!
			</button>
		</form>
	);
}

const DisplayPalette = ({ palette }) => {
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
		<div className="display-palette">
			{
				blocks
			}
		</div>
	);
}

export default ExtractionBlock;
